package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nilslice/protolock"
)

const usage = `Track your .proto files and prevent changes to messages and services which impact API compatibilty.

Copyright: Steve Manuel <nilslice@gmail.com>
Released under the BSD-3-Clause license.

Usage:
	protolock <command> [options]

Commands:
	-h, --help, help	display the usage information for protolock
	init			initialize a proto.lock file from current tree
	status			check for breaking changes and report conflicts
	commit			overwrite proto.lock file with current tree

Options:
	--strict [true]		enable strict mode and enforce all built-in rules
	--debug	[false]		enable debug mode and output debug messages
	--ignore 		comma-separated list of filepaths to ignore
`

var (
	options = flag.NewFlagSet("options", flag.ExitOnError)
	debug   = options.Bool("debug", false, "toggle debug mode for verbose output")
	strict  = options.Bool("strict", true, "toggle strict mode, to determine which rules are enforced")
	ignore  = options.String("ignore", "", "comma-separated list of filepaths to ignore")
)

func main() {
	// exit if no command (i.e. help, -h, --help, init, status, or commit)
	if len(os.Args) < 2 {
		fmt.Println(usage)
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

		err = saveToLockFile(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "status":
		report, err := protolock.Status(*ignore)
		if err != nil {
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

			fmt.Println(err)
		}

	default:
		os.Exit(0)
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
