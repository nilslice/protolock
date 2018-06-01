package protolock

import (
	"errors"
	"os"
)

// Status will report on any issues encountered when comparing the updated tree
// of parsed proto files and the current proto.lock file.
func Status(ignore string) (Report, error) {
	updated, err := getUpdatedLock(ignore)
	if err != nil {
		return Report{}, err
	}

	lockFile, err := openLockFile()
	if err != nil {
		if os.IsNotExist(err) {
			msg := `no "proto.lock" file found, first run "init"`
			return Report{}, errors.New(msg)
		}
		return Report{}, err
	}
	defer lockFile.Close()

	current, err := protolockFromReader(lockFile)
	if err != nil {
		return Report{}, err
	}

	return compare(current, *updated)
}
