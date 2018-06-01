package protolock

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gpfPath = filepath.Join("testdata", "getProtoFiles")

func TestGetProtoFilesFiltersDirectories(t *testing.T) {
	files, err := getProtoFiles(gpfPath, "")
	require.NoError(t, err)

	path := filepath.Join(gpfPath, "directory.proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "include.proto")
	assert.Contains(t, files, path)
}

func TestGetProtoFilesFiltersNonProto(t *testing.T) {
	files, err := getProtoFiles(gpfPath, "")
	require.NoError(t, err)

	path := filepath.Join(gpfPath, "directory.proto", "test.non-proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "include.proto")
	assert.Contains(t, files, path)
}

func TestGetProtoFilesIgnoresDirectories(t *testing.T) {
	files, err := getProtoFiles(gpfPath, "exclude")
	require.NoError(t, err)

	path := filepath.Join(gpfPath, "exclude", "test.proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "include.proto")
	assert.Contains(t, files, path)
}

func TestGetProtoFilesIgnoresFiles(t *testing.T) {
	files, err := getProtoFiles(gpfPath, filepath.Join("include", "exclude.proto"))
	require.NoError(t, err)

	path := filepath.Join(gpfPath, "include", "exclude.proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "include.proto")
	assert.Contains(t, files, path)
}

func TestGetProtoFilesIgnoresMultiple(t *testing.T) {
	paths := []string{"exclude", filepath.Join("include", "exclude.proto")}
	ignores := strings.Join(paths, ",")
	files, err := getProtoFiles(gpfPath, ignores)
	require.NoError(t, err)

	path := filepath.Join(gpfPath, "exclude", "test.proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "exclude.proto")
	assert.NotContains(t, files, path)

	path = filepath.Join(gpfPath, "include", "include.proto")
	assert.Contains(t, files, path)
}
