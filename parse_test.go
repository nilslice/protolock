package protolock

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const protoWithImports = `
syntax = "proto3";

import "testdata/test.proto"

package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

const protoWithOptions = `
syntax = "proto3";

package test;

message Channel {
  option (ext.persisted) = true
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

var gpfPath = filepath.Join("testdata", "getProtoFiles")


func TestParseIncludingImports(t *testing.T) {
	r := strings.NewReader(protoWithImports)

	entry, _ := parse(r)

	assert.Equal(t, "testdata/test.proto", entry.Imports[0].Path)
}

func TestParseIncludingOptions(t *testing.T) {
	r := strings.NewReader(protoWithOptions)

	entry, _ := parse(r)

	assert.Equal(t, "(ext.persisted)", entry.Messages[0].Options[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Options[0].Value)
}

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
