package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nilslice/protolock"
)

var (
	debug      = flag.Bool("debug", false, "toggle debug mode for verbose output")
	strictMode = flag.Bool("strict", true, "toggle strict mode, to determine which rules are enforced")
)

func main() {
	flag.Parse()

	// XXX: currently here as placeholder until better CLI implementation
	// is completed. This includes debug and strictMode vars in block above.
	protolock.SetDebug(*debug)
	protolock.SetStrictMode(*strictMode)

	if len(os.Args) < 2 {
		os.Exit(0)
	}

	switch os.Args[1] {
	case "init":
		r, err := protolock.Init()
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
		r, err := protolock.Commit()
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
		report, err := protolock.Status()
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
		fmt.Println(os.Args)
	}
}

func saveToLockFile(r io.Reader) error {
	lockfile, err := os.Create("proto.lock")
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
