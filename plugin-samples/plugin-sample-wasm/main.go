package main

import (
	"encoding/json"

	pdk "github.com/extism/go-pdk"
	"github.com/nilslice/protolock"
	"github.com/nilslice/protolock/extend"
)

// an Extism plugin uses a 'PDK' to communicate data input and output from its host system, in
// this case, the `protolock` command.

// see https://extism.org and https://github.com/extism/extism for more information.

// In order to satisfy the current usage, an Extism Protolock plugin must export a single function
// "status" with the following signature:

//export status
func status() int32 {
	// rather than taking input from stdin, like native Protolock plugins, Extism plugins take data
	// from their host, using the `pdk.Input()` function, returning bytes from protolock.
	var data extend.Data
	err := json.Unmarshal(pdk.Input(), &data.Current)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	// with the `extend.Data` available, you would do some checks on the current and updated set of
	// `proto.lock` representations. Here we are adding a warning to demonstrate that the plugin
	// works with some known data output to verify.
	warning := protolock.Warning{
		Filepath: "fake.proto",
		Message:  "An Extism plugin ran and checked the status of the proto.lock files",
		RuleName: "RuleNameXYZ",
	}
	data.PluginWarnings = append(data.PluginWarnings, warning)

	b, err := json.Marshal(data)
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	// tather than writing data to stdout, like native Protolock plugins, Extism plugins provide
	// data back to their host, using the `pdk.Output()` function, returning bytes to protolock.
	pdk.Output(b)

	// non-zero return code here will result in Extism detecting an error.
	return 0
}

// this Go code is compiled to WebAssembly, and current compilers expect some entrypoint, even if
// this function isn't called.
func main() {}
