package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/nilslice/protolock"
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
	--plugins 		comma-separated list of executable protolock plugin names
	--lockdir [.]		directory of proto.lock file
	--protoroot [.]		root of directory tree containing proto files
`

var (
	options   = flag.NewFlagSet("options", flag.ExitOnError)
	debug     = options.Bool("debug", false, "toggle debug mode for verbose output")
	strict    = options.Bool("strict", true, "enable strict mode and enforce all built-in rules")
	ignore    = options.String("ignore", "", "comma-separated list of filepaths to ignore")
	force     = options.Bool("force", false, "force commit to rewrite proto.lock file and disregard warnings")
	plugins   = options.String("plugins", "", "comma-separated list of executable protolock plugin names")
	lockDir   = options.String("lockdir", ".", "directory of proto.lock file")
	protoRoot = options.String("protoroot", ".", "root of directory tree containing proto files")
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

	cfg, err := protolock.NewConfig(*lockDir, *protoRoot, *ignore)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// switch through known commands
	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Println(usage)

	case "init":
		r, err := protolock.Init(*cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = saveToLockFile(*cfg, r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "commit":
		r, err := protolock.Commit(*cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// if force option is false (default), then disallow commit if
		// there are any warnings encountered by runing a status check.
		if !*force {
			report, err := protolock.Status(*cfg)
			if err != nil {
				handleReport(report, err)
			}

			if len(report.Warnings) > 0 {
				os.Exit(1)
			}
		}

		err = saveToLockFile(*cfg, r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "status":
		report, err := protolock.Status(*cfg)
		if err != protolock.ErrWarningsFound && err != nil {
			fmt.Println("[protolock]:", err)
			os.Exit(1)
		}

		// if plugins are provided, attempt to execute each as a exeutable
		// located in the user's OS executable path as reported by stdlib's
		// exec.LookPath func
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

func handleReport(report *protolock.Report, err error) {
	if len(report.Warnings) > 0 {

		orderByPathAndMessage(report.Warnings)

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

func orderByPathAndMessage(warnings []protolock.Warning) {
	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].Filepath < warnings[j].Filepath {
			return true
		}
		if warnings[i].Filepath > warnings[j].Filepath {
			return false
		}
		return warnings[i].Message < warnings[j].Message
	})
}

func saveToLockFile(cfg protolock.Config, r io.Reader) error {
	lockfile, err := os.Create(cfg.LockFilePath())
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
