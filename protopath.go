package protolock

import (
	"path/filepath"
	"strings"
)

const (
	filesep  = string(filepath.Separator)
	protosep = ":/:"
)

type Protopath string

// convert a path in the Protopath format to the OS path format
func OSPath(ProtoPath Protopath) Protopath {
	return Protopath(
		strings.Replace(string(ProtoPath), protosep, filesep, -1),
	)
}

// convert a path in the OS path format to Protopath format
func ProtoPath(OSPath Protopath) Protopath {
	return Protopath(
		strings.Replace(string(OSPath), filesep, protosep, -1),
	)
}

func (p Protopath) String() string {
	return string(p)
}
