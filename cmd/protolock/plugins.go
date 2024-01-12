package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	extism "github.com/extism/go-sdk"
	"github.com/nilslice/protolock"
	"github.com/nilslice/protolock/extend"
)

const logPrefix = "[protolock]"

func runPlugins(
	pluginList string,
	report *protolock.Report,
	debug bool,
) (*protolock.Report, error) {
	inputData := &bytes.Buffer{}

	err := json.NewEncoder(inputData).Encode(&extend.Data{
		Current:           report.Current,
		Updated:           report.Updated,
		ProtolockWarnings: report.Warnings,
		PluginWarnings:    []protolock.Warning{},
	})
	if err != nil {
		return nil, err
	}

	// collect plugin warnings and errors as they are returned from plugins
	pluginWarningsChan := make(chan []protolock.Warning)
	pluginsDone := make(chan struct{})
	pluginErrsChan := make(chan error)
	var allPluginErrors []error
	go func() {
		for {
			select {
			case <-pluginsDone:
				return

			case err := <-pluginErrsChan:
				if err != nil {
					allPluginErrors = append(allPluginErrors, err)
				}

			case warnings := <-pluginWarningsChan:
				for _, warning := range warnings {
					report.Warnings = append(report.Warnings, warning)
				}
			}
		}
	}()

	wg := &sync.WaitGroup{}
	plugins := strings.Split(pluginList, ",")
	for _, name := range plugins {
		wg.Add(1)

		// copy input data to be passed in to and processed by each plugin
		pluginInputData := bytes.NewReader(inputData.Bytes())

		// run all provided plugins in parallel, each recieving the same current
		// and updated Protolock structs from the `protolock status` call
		go func(name string) {
			defer wg.Done()
			// output is populated either by the execution of an Extism plugin or a native binary
			var output []byte
			name = strings.TrimSpace(name)
			path := name

			if debug {
				fmt.Println(logPrefix, name, "running plugin")
			}

			if strings.HasSuffix(name, ".wasm") {
				// do extism call
				manifest := extism.Manifest{
					Wasm: []extism.Wasm{extism.WasmFile{Path: name}},
					// TODO: consider enabling external configuration to add hosts and paths
					// AllowedHosts: []string{},
					// AllowedPaths: map[string]string{},
				}

				plugin, err := extism.NewPlugin(context.Background(), manifest, extism.PluginConfig{EnableWasi: true}, nil)
				if err != nil {
					fmt.Println(logPrefix, name, "failed to create extism plugin:", err)
					return
				}

				var exitCode uint32
				exitCode, output, err = plugin.Call("status", inputData.Bytes())
				if err != nil {
					fmt.Println(logPrefix, name, "plugin exec error: ", err, "code:", exitCode)
					pluginErrsChan <- wrapPluginErr(name, path, err, output)
					return
				}

			} else {
				path, err = exec.LookPath(name)
				if err != nil {
					if path == "" {
						path = name
					}
					fmt.Println(logPrefix, name, "plugin exec error:", err)
					return
				}

				// initialize the executable to be called from protolock using the
				// absolute path and copy of the input data
				plugin := &exec.Cmd{
					Path:  path,
					Stdin: pluginInputData,
				}

				// execute the plugin and capture the output
				output, err = plugin.CombinedOutput()
				if err != nil {
					pluginErrsChan <- wrapPluginErr(name, path, err, output)
					return
				}
			}

			pluginData := &extend.Data{}
			err = json.Unmarshal(output, pluginData)
			if err != nil {
				fmt.Println(logPrefix, name, "plugin data decode error:", err)
				// TODO: depending on the plugin, "output" could be quite
				// verbose, though may not warrant debug flag guard.
				if debug {
					fmt.Println(
						logPrefix, name, "plugin output:", string(output),
					)
				}
				return
			}

			// gather all warnings from each plugin, and send to warning chan
			// collector as a slice to keep together
			if pluginData.PluginWarnings != nil {
				pluginWarningsChan <- pluginData.PluginWarnings
			}

			if pluginData.PluginErrorMessage != "" {
				pluginErrsChan <- wrapPluginErr(
					name,
					path,
					errors.New(pluginData.PluginErrorMessage),
					output,
				)
			}
		}(name)
	}

	wg.Wait()
	pluginsDone <- struct{}{}

	if allPluginErrors != nil {
		var errorMsgs []string
		for _, pluginError := range allPluginErrors {
			errorMsgs = append(errorMsgs, pluginError.Error())
		}

		return nil, fmt.Errorf(
			"accumulated plugin errors: \n%s",
			strings.Join(errorMsgs, "\n"),
		)
	}

	return report, nil
}

func wrapPluginErr(name, path string, err error, output []byte) error {
	return fmt.Errorf(
		"%s (%s): %v\n%s",
		name, path, err, strings.ReplaceAll(
			string(output),
			protolock.ProtoSep, protolock.FileSep,
		),
	)
}
