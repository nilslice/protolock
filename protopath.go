package protolock

import (
	"path/filepath"
	"strings"
)

const (
	filesep  = string(filepath.Separator)
	protosep = ":/:"
)

type protopath string

// convert a path in the protopath format to the OS path format
func osPath(protoPath protopath) protopath {
	return protopath(
		strings.Replace(string(protoPath), protosep, filesep, -1),
	)
}

// convert a path in the OS path format to protopath format
func protoPath(osPath protopath) protopath {
	return protopath(
		strings.Replace(string(osPath), filesep, protosep, -1),
	)
}

func (p protopath) String() string {
	return string(p)
}
