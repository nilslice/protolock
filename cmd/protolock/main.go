package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nilslice/protolock"
)

func main() {
	flag.Parse()

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

	case "status":
		report, err := protolock.Status()
		if err != nil {
			if len(report.Warnings) > 0 {
				for _, w := range report.Warnings {
					fmt.Fprintf(
						os.Stdout,
						"%s [%s]\n",
						w.Message, w.Filepath,
					)
				}

				term := "issue"
				if len(report.Warnings) > 1 {
					term = "issues"
				}
				fmt.Fprintf(
					os.Stdout,
					"Encountered %d %s during analysis.\n",
					len(report.Warnings), term,
				)
				os.Exit(1)
			}

			fmt.Println(err)
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
