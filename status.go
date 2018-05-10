package protolock

// Status will report on any issues encountered when comparing the updated tree
// of parsed proto files and the current proto.lock file.
func Status() (Report, error) {
	updated, err := getUpdatedLock()
	if err != nil {
		return Report{}, err
	}

	lockFile, err := openLockFile()
	if err != nil {
		return Report{}, err
	}
	defer lockFile.Close()

	current, err := protolockFromReader(lockFile)
	if err != nil {
		return Report{}, err
	}

	return compare(current, *updated)
}
