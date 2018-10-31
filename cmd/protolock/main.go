package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nilslice/protolock"
	"github.com/nilslice/protolock/extend"
)

const info = `Track your .proto files and prevent changes to messages and services which impact API compatibilty.

Copyright Steve Manuel <nilslice@gmail.com>
Released under the BSD-3-Clause license.
`

const usage = `
Usage:
	protolock <command> [options]

Commands:
	-h, --help, help	display the usage information for protolock
	init			initialize a proto.lock file from current tree
	status			check for breaking changes and report conflicts
	commit			rewrite proto.lock file with current tree if no conflicts (--force to override)

Options:
	--strict [true]		enable strict mode and enforce all built-in rules
	--debug	[false]		enable debug mode and output debug messages
	--ignore 		comma-separated list of filepaths to ignore
	--force [false]		forces commit to rewrite proto.lock file and disregards warnings
	--plugins			comma-separated list of executable protolock plugin names
`

var (
	options = flag.NewFlagSet("options", flag.ExitOnError)
	debug   = options.Bool("debug", false, "toggle debug mode for verbose output")
	strict  = options.Bool("strict", true, "enable strict mode and enforce all built-in rules")
	ignore  = options.String("ignore", "", "comma-separated list of filepaths to ignore")
	force   = options.Bool("force", false, "force commit to rewrite proto.lock file and disregard warnings")
	plugins = options.String("plugins", "", "comma-separated list of executable protolock plugin names")
)

func main() {
	// exit if no command (i.e. help, -h, --help, init, status, or commit)
	if len(os.Args) < 2 {
		fmt.Println(info + usage)
		os.Exit(0)
	}

	// parse and set options flags
	options.Parse(os.Args[2:])
	protolock.SetDebug(*debug)
	protolock.SetStrict(*strict)

	// switch through known commands
	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Println(usage)

	case "init":
		r, err := protolock.Init(*ignore)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = saveToLockFile(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "commit":
		r, err := protolock.Commit(*ignore)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// if force option is false (default), then disallow commit if
		// there are any warnings encountered by runing a status check.
		if !*force {
			report, err := protolock.Status(*ignore)
			if err != nil {
				handleReport(report, err)
			}

			if len(report.Warnings) > 0 {
				os.Exit(1)
			}
		}

		err = saveToLockFile(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "status":
		report, err := protolock.Status(*ignore)
		if err != protolock.ErrWarningsFound && err != nil {
			fmt.Println("[protolock]:", err)
			os.Exit(1)
		}

		// if plugins are provided, attempt to execute each as a binary located
		// in the user's OS executable filepath using exec.LookPath
		if *plugins != "" {
			report, err = runPlugins(*plugins, report)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		handleReport(report, err)

	default:
		os.Exit(0)
	}
}

func runPlugins(pluginList string, report *protolock.Report) (*protolock.Report, error) {
	inputData := &bytes.Buffer{}

	err := json.NewEncoder(inputData).Encode(&extend.Data{
		Current:          report.Current,
		Updated:          report.Updated,
		ExistingWarnings: report.Warnings,
	})
	if err != nil {
		return nil, err
	}

	plugins := strings.Split(pluginList, ",")
	for i, name := range plugins {
		name = strings.TrimSpace(name)
		path, err := exec.LookPath(name)
		if err != nil {
			return wrapPluginErr(path, err)
		}

		plugin := &exec.Cmd{
			Path:  path,
			Stdin: inputData,
		}

		// execute the plugin and capture the output
		output, err := plugin.Output()
		if err != nil {
			return wrapPluginErr(path, err)
		}

		// reset inputData before writing the output to it from previously
		// executed plugin
		inputData.Reset()

		// if this is not the last plugin to execute, save the output of the
		// previous plugin to be passed into the next plugin
		if i != len(plugins)-1 {
			_, err := inputData.Write(output)
			if err != nil {
				return wrapPluginErr(path, err)
			}
		} else {
			pluginData := &extend.Data{}
			err = json.Unmarshal(output, pluginData)
			if err != nil {
				return nil, err
			}

			report.Warnings = pluginData.ExistingWarnings
		}
	}

	return report, nil
}

func handleReport(report *protolock.Report, err error) {
	if len(report.Warnings) > 0 {
		for _, w := range report.Warnings {
			fmt.Fprintf(
				os.Stdout,
				"CONFLICT: %s [%s]\n",
				w.Message, w.Filepath,
			)
		}
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func saveToLockFile(r io.Reader) error {
	lockfile, err := os.Create(protolock.LockFileName)
	if err != nil {
		return err
	}
	defer lockfile.Close()

	_, err = io.Copy(lockfile, r)
	if err != nil {
		return err
	}

	return nil
}

func wrapPluginErr(path string, err error) (*protolock.Report, error) {
	return nil, fmt.Errorf("[protolock:plugin] %v (%s)", err, path)
}
