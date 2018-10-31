package main

import (
	"github.com/nilslice/protolock"
	"github.com/nilslice/protolock/extend"
)

func main() {
	plugin := extend.NewPlugin("sample") // "sample" is arbitrary name used to correlate error messages
	plugin.Init(func(data *extend.Data) (*extend.Data, error) {
		warnings := AddWarningsForExample(data.Current, data.Updated)
		data.PluginWarnings = append(data.PluginWarnings, warnings...)
		return data, nil
	})
}

func AddWarningsForExample(cur, upd protolock.Protolock) []protolock.Warning {
	return []protolock.Warning{
		{Filepath: protolock.Protopath(""), Message: "A sample warning!"},
		{Filepath: protolock.Protopath(""), Message: "Another sample warning.. ah!"},
	}
}
