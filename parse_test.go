package protolock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProtoFilesFiltersDirectories(t *testing.T) {
	files, err := getProtoFiles("testdata/getProtoFiles", "")
	assert.NoError(t, err)

	assert.NotContains(t, files, "testdata/getProtoFiles/directory.proto")
	assert.Contains(t, files, "testdata/getProtoFiles/include/include.proto")
}

func TestGetProtoFilesFiltersNonProto(t *testing.T) {
	files, err := getProtoFiles("testdata/getProtoFiles", "")
	assert.NoError(t, err)

	assert.NotContains(t, files, "testdata/getProtoFiles/directory.proto/test.non-proto")
	assert.Contains(t, files, "testdata/getProtoFiles/include/include.proto")
}

func TestGetProtoFilesIgnoresDirectories(t *testing.T) {
	files, err := getProtoFiles("testdata/getProtoFiles", "exclude")
	assert.NoError(t, err)

	assert.NotContains(t, files, "testdata/getProtoFiles/exclude/test.proto")
	assert.Contains(t, files, "testdata/getProtoFiles/include/include.proto")
}

func TestGetProtoFilesIgnoresFiles(t *testing.T) {
	files, err := getProtoFiles("testdata/getProtoFiles", "include/exclude.proto")
	assert.NoError(t, err)

	assert.NotContains(t, files, "testdata/getProtoFiles/include/exclude.proto")
	assert.Contains(t, files, "testdata/getProtoFiles/include/include.proto")
}

func TestGetProtoFilesIgnoresMultiple(t *testing.T) {
	files, err := getProtoFiles("testdata/getProtoFiles", "exclude,include/exclude.proto")
	assert.NoError(t, err)

	assert.NotContains(t, files, "testdata/getProtoFiles/exclude/test.proto")
	assert.NotContains(t, files, "testdata/getProtoFiles/include/exclude.proto")
	assert.Contains(t, files, "testdata/getProtoFiles/include/include.proto")
}
