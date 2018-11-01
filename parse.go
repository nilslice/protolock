package protolock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emicklei/proto"
)

const LockFileName = "proto.lock"

type Protolock struct {
	Definitions []Definition `json:"definitions,omitempty"`
}

type Definition struct {
	Filepath Protopath `json:"protopath,omitempty"`
	Def      Entry     `json:"def,omitempty"`
}

type Entry struct {
	Enums    []Enum    `json:"enums,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Services []Service `json:"services,omitempty"`
	Imports  []Import  `json:"imports,omitempty"`
}

type Import struct {
	Path string `json:"path,omitempty"`
}

type Option struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Message struct {
	Name          string    `json:"name,omitempty"`
	Fields        []Field   `json:"fields,omitempty"`
	Maps          []Map     `json:"maps,omitempty"`
	ReservedIDs   []int     `json:"reserved_ids,omitempty"`
	ReservedNames []string  `json:"reserved_names,omitempty"`
	Filepath      Protopath `json:"filepath,omitempty"`
	Messages      []Message `json:"messages,omitempty"`
	Options       []Option  `json:"options,omitempty"`
}

type EnumField struct {
	Name    string `json:"name,omitempty"`
	Integer int    `json:"integer"`
}

type Enum struct {
	Name          string      `json:"name,omitempty"`
	EnumFields    []EnumField `json:"enum_fields,omitempty"`
	ReservedIDs   []int       `json:"reserved_ids,omitempty"`
	ReservedNames []string    `json:"reserved_names,omitempty"`
	AllowAlias    bool        `json:"allow_alias,omitempty"`
}

type Map struct {
	KeyType string `json:"key_type,omitempty"`
	Field   Field  `json:"field,omitempty"`
}

type Field struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	IsRepeated bool   `json:"is_repeated,omitempty"`
}

type Service struct {
	Name     string    `json:"name,omitempty"`
	RPCs     []RPC     `json:"rpcs,omitempty"`
	Filepath Protopath `json:"filepath,omitempty"`
}

type RPC struct {
	Name        string `json:"name,omitempty"`
	InType      string `json:"in_type,omitempty"`
	OutType     string `json:"out_type,omitempty"`
	InStreamed  bool   `json:"in_streamed,omitempty"`
	OutStreamed bool   `json:"out_streamed,omitempty"`
}

type Report struct {
	Current, Updated Protolock `json:"current,omitempty"`
	Warnings         []Warning `json:"warnings,omitempty"`
}

type Warning struct {
	Filepath Protopath `json:"filepath,omitempty"`
	Message  string    `json:"message,omitempty"`
}

type ProtoFile struct {
	ProtoPath Protopath
	Entry     Entry
}

var (
	enums []Enum
	msgs  []Message
	svcs  []Service
	imps  []Import

	ErrWarningsFound = errors.New("comparison found one or more warnings")
)

func parse(r io.Reader) (Entry, error) {
	parser := proto.NewParser(r)
	def, err := parser.Parse()
	if err != nil {
		return Entry{}, err
	}

	enums = []Enum{}
	msgs = []Message{}
	svcs = []Service{}
	imps = []Import{}

	proto.Walk(
		def,
		proto.WithEnum(withEnum),
		proto.WithService(withService),
		proto.WithMessage(withMessage),
		protoWithImport(withImport),
	)

	return Entry{
		Enums:    enums,
		Messages: msgs,
		Services: svcs,
		Imports:  imps,
	}, nil
}

func withEnum(e *proto.Enum) {
	errs := checkComments(e)
	if errs != nil {
		for _, err := range errs {
			switch err {
			case ErrSkipEntry:
				return
			}
		}
	}

	// handle nested enum within message, prepend message name to enum name
	if p, ok := e.Parent.(*proto.Message); ok {
		if p != nil {
			e.Name = fmt.Sprintf("%s.%s", p.Name, e.Name)
		}
	}

	enums = append(enums, parseEnum(e))
}

func parseEnum(e *proto.Enum) Enum {
	enum := Enum{
		Name: e.Name,
	}

	for _, v := range e.Elements {
		if e, ok := v.(*proto.EnumField); ok {
			enum.EnumFields = append(enum.EnumFields, EnumField{
				Name:    e.Name,
				Integer: e.Integer,
			})
		}

		if r, ok := v.(*proto.Reserved); ok {
			// collect all reserved field IDs from the ranges
			for _, rng := range r.Ranges {
				// if range is only a single value, skip loop and
				// append single value to message's reserved slice
				if rng.From == rng.To {
					enum.ReservedIDs = append(enum.ReservedIDs, rng.From)
					continue
				}
				// add each item from the range inclusively
				for id := rng.From; id <= rng.To; id++ {
					enum.ReservedIDs = append(enum.ReservedIDs, id)
				}
			}

			// add all reserved field names
			enum.ReservedNames = append(enum.ReservedNames, r.FieldNames...)
		}
	}

	return enum
}

func withService(s *proto.Service) {
	errs := checkComments(s)
	if errs != nil {
		for _, err := range errs {
			switch err {
			case ErrSkipEntry:
				return
			}
		}
	}

	svc := Service{
		Name: s.Name,
	}

	for _, v := range s.Elements {
		if r, ok := v.(*proto.RPC); ok {
			svc.RPCs = append(svc.RPCs, RPC{
				Name:        r.Name,
				InType:      r.RequestType,
				OutType:     r.ReturnsType,
				InStreamed:  r.StreamsRequest,
				OutStreamed: r.StreamsReturns,
			})
		}
	}

	svcs = append(svcs, svc)
}

func withMessage(m *proto.Message) {
	errs := checkComments(m)
	if errs != nil {
		for _, err := range errs {
			switch err {
			case ErrSkipEntry:
				return
			}
		}
	}

	if _, ok := m.Parent.(*proto.Proto); !ok {
		return
	}

	msgs = append(msgs, parseMessage(m))
}

func parseMessage(m *proto.Message) Message {
	msg := Message{
		Name: m.Name,
	}

	for _, v := range m.Elements {

		if f, ok := v.(*proto.NormalField); ok {
			msg.Fields = append(msg.Fields, Field{
				ID:         f.Sequence,
				Name:       f.Name,
				Type:       f.Type,
				IsRepeated: f.Repeated,
			})
		}

		if mp, ok := v.(*proto.MapField); ok {
			f := mp.Field
			msg.Maps = append(msg.Maps, Map{
				KeyType: mp.KeyType,
				Field: Field{
					ID:         f.Sequence,
					Name:       f.Name,
					Type:       f.Type,
					IsRepeated: false,
				},
			})
		}

		if oo, ok := v.(*proto.Oneof); ok {
			var fields []Field
			for _, el := range oo.Elements {
				if f, ok := el.(*proto.OneOfField); ok {
					fields = append(fields, Field{
						ID:         f.Sequence,
						Name:       f.Name,
						Type:       f.Type,
						IsRepeated: false,
					})
				}
			}
			msg.Fields = append(msg.Fields, fields...)
		}

		if r, ok := v.(*proto.Reserved); ok {
			// collect all reserved field IDs from the ranges
			for _, rng := range r.Ranges {
				// if range is only a single value, skip loop and
				// append single value to message's reserved slice
				if rng.From == rng.To {
					msg.ReservedIDs = append(msg.ReservedIDs, rng.From)
					continue
				}
				// add each item from the range inclusively
				for id := rng.From; id <= rng.To; id++ {
					msg.ReservedIDs = append(msg.ReservedIDs, id)
				}
			}

			// add all reserved field names
			msg.ReservedNames = append(msg.ReservedNames, r.FieldNames...)
		}

		if o, ok := v.(*proto.Option); ok {
			msg.Options = append(msg.Options, Option{
				Name:  o.Name,
				Value: o.Constant.Source,
			})
		}

		if m, ok := v.(*proto.Message); ok {
			msg.Messages = append(msg.Messages, parseMessage(m))
		}
	}

	return msg
}

func protoWithImport(apply func(p *proto.Import)) proto.Handler {
	return func(v proto.Visitee) {
		if s, ok := v.(*proto.Import); ok {
			apply(s)
		}
	}
}

func withImport(im *proto.Import) {
	imp := Import{
		Path: im.Filename,
	}
	imps = append(imps, imp)
}

// openLockFile opens and returns the lock file on disk for reading.
func openLockFile() (io.ReadCloser, error) {
	f, err := os.Open(LockFileName)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// protolockFromReader unmarshals a proto.lock file into a Protolock struct.
func protolockFromReader(r io.Reader) (Protolock, error) {
	buf := bytes.Buffer{}
	_, err := io.Copy(&buf, r)
	if err != nil {
		return Protolock{}, err
	}

	var lock Protolock
	err = json.Unmarshal(buf.Bytes(), &lock)
	if err != nil {
		return Protolock{}, err
	}

	return lock, nil
}

// compare returns a Report struct and an error which indicates that there is
// one or more warnings to report to the caller. If no error is returned, the
// Report can be ignored.
func compare(current, update Protolock) (*Report, error) {
	var warnings []Warning
	var wg sync.WaitGroup
	report := &Report{
		Current: current,
		Updated: update,
	}
	for _, fn := range ruleFuncs {
		wg.Add(1)
		go func() {
			if w, ok := fn(current, update); !ok {
				warnings = append(warnings, w...)
			}
			wg.Done()
		}()
		wg.Wait()
	}
	report.Warnings = warnings

	if len(report.Warnings) != 0 {
		return report, ErrWarningsFound
	}

	return report, nil
}

// getProtoFiles finds recursively all .proto files to be processed.
func getProtoFiles(root string, ignores string) ([]string, error) {
	protoFiles := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if not a .proto file, do not attempt to parse.
		if !strings.HasSuffix(info.Name(), protoSuffix) {
			return nil
		}

		// skip to next if is a directory
		if info.IsDir() {
			return nil
		}

		// skip if path is within an ignored path
		if ignores != "" {
			for _, ignore := range strings.Split(ignores, ",") {
				rel, err := filepath.Rel(filepath.Join(root, ignore), path)
				if err != nil {
					return nil
				}

				if !strings.HasPrefix(rel, "../") {
					return nil
				}
			}
		}

		protoFiles = append(protoFiles, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return protoFiles, nil
}

// getUpdatedLock finds all .proto files recursively in tree, parse each file
// and accumulate all definitions into an updated Protolock.
func getUpdatedLock(ignores string) (*Protolock, error) {
	// files is a slice of struct `ProtoFile` to be joined into the proto.lock file.
	var files []ProtoFile

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	protoFiles, err := getProtoFiles(root, ignores)
	if err != nil {
		return nil, err
	}

	for _, path := range protoFiles {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		entry, err := parse(f)
		if err != nil {
			printIfErr(f.Close())
			return nil, err
		}

		localPath := strings.TrimPrefix(path, root)
		localPath = strings.TrimPrefix(localPath, string(filepath.Separator))
		protoFile := ProtoFile{
			ProtoPath: ProtoPath(Protopath(localPath)),
			Entry:     entry,
		}
		files = append(files, protoFile)

		// manually close the file to prevent `too many open files` error
		printIfErr(f.Close())
	}

	// add all the definitions from the updated set of protos to a Protolock
	// used for analysis and comparison against the current Protolock, saved
	// as the proto.lock file in the current directory
	var updated Protolock
	for _, file := range files {
		updated.Definitions = append(updated.Definitions, Definition{
			Filepath: file.ProtoPath,
			Def:      file.Entry,
		})
	}

	return &updated, nil
}

func printIfErr(err error) {
	if err != nil {
		fmt.Printf("protolock: %v\n", err)
	}
}
