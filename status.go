package protolock

import (
	"os"
	"errors"
)

// Status will report on any issues encountered when comparing the updated tree
// of parsed proto files and the current proto.lock file.
func Status(ignore string) (Report, error) {
	return LoadProtolocksAndCheck(ignore, predefinedRuleFuncs)
}

// Loads current and updated protolock and validate if the changes pass the rules
func LoadProtolocksAndCheck(ignore string, ruleFuncs []RuleFunc) (Report, error) {
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

	return compare(current, *updated, ruleFuncs)
}
