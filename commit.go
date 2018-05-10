package protolock

import (
	"fmt"
	"io"
	"os"
)

// Commit will return an io.Reader with the lock representation data for caller to
// use as needed.
func Commit() (io.Reader, error) {
	if _, err := os.Stat(lockFileName); err != nil && os.IsNotExist(err) {
		fmt.Println(`no proto.lock found, please run "init" first`)
		os.Exit(1)
	}
	updated, err := getUpdatedLock()
	if err != nil {
		return nil, err
	}

	return readerFromProtolock(updated)
}
