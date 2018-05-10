package protolock

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const protoSuffix = ".proto"

// Init will return an io.Reader with the lock representation data for caller to
// use as needed.
func Init() (io.Reader, error) {
	if _, err := os.Stat(lockFileName); err == nil && !os.IsNotExist(err) {
		fmt.Println(`a proto.lock file was already found, use "commit" to update`)
		os.Exit(1)
	}
	updated, err := getUpdatedLock()
	if err != nil {
		return nil, err
	}

	return readerFromProtolock(updated)
}

func readerFromProtolock(lock *Protolock) (io.Reader, error) {
	b, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(b)), nil
}
