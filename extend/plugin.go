package extend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nilslice/protolock"
)

// Data contains the current and updated Protolock structs created by the
// `protolock` internal parser and deserializer, and a slice of Warning structs
// from the initial `protolock status`, and subsequent plugin invocations.
type Data struct {
	Current, Updated protolock.Protolock
	ExistingWarnings []protolock.Warning
}

// PluginFunc is a function which defines plugin behavior, and is provided a
// pointer to Data.
type PluginFunc func(d *Data) (*Data, error)

type plugin struct {
	name string
}

// NewPlugin returns a plugin instance for a plugin to be initialized.
func NewPlugin(name string) *plugin {
	return &plugin{
		name: name,
	}
}

// Init is called by plugin code and is provided a PluginFunc from the caller
// to handle the input Data (read from stdin).
func (p *plugin) Init(fn PluginFunc) {
	// read from stdin to get serialized bytes
	input := &bytes.Buffer{}
	_, err := io.Copy(input, os.Stdin)
	if err != nil {
		p.wrapErrAndLog(err)
	}

	// deserialize bytes into *Data
	inputData := &Data{}
	err = json.Unmarshal(input.Bytes(), inputData)
	if err != nil {
		p.wrapErrAndLog(err)
	}

	// execute "fn" and pass it the *Data, where the plugin would read and
	// compare the current and updated Protolock values and append custom
	// Warnings for their own defined rules
	outputData, err := fn(inputData)
	if err != nil {
		p.wrapErrAndLog(err)
	}
	outputData.Current = inputData.Current
	outputData.Updated = inputData.Updated

	// serialize *Data back and write to stdout
	p.wrapErrAndLog(json.NewEncoder(os.Stdout).Encode(outputData))

}

func (p *plugin) wrapErrAndLog(err error) {
	fmt.Fprintf(os.Stderr, "[protolock:plugin] %s: %v", p.name, err)
}
