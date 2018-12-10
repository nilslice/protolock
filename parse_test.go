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

import "testdata/test.proto";

package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

const protoWithMessageOptions = `
syntax = "proto3";

package test;

message Channel {
  option (ext.persisted) = true;
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

const protoWithNestedMessageOptions = `
syntax = "proto3";

package test;

message Channel {
  option (ext.persisted) = { opt1: true opt2: false };
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

const protoWithFieldOptions = `
syntax = "proto3";

package test;

message Channel {
  int64 id = 1;
  string name = 2 [(personal) = true, (owner) = 'test'];
  string description = 3;
  map<string, int32> attributes = 4 [(personal) = true];
}
`

const protoWithNestedFieldOptions = `
syntax = "proto3";

package test;

message Channel {
 int64 id = 1;
 string name = 2;
 string description = 3;
 map<string, int32> attributes = 4;
 string address = 5 [(custom_options).personal = true, (custom_options).internal = false];
}
`

const protoWithNestedFieldOptionsAggregated = `
syntax = "proto3";

package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3 [(custom_options_commas) = { personal: true, internal: false, owner: "some owner" }];
  map<string, int32> attributes = 4;
  string address = 5 [(custom_options) = { personal: true internal: false owner: "some owner" }];
}
`

const protoWithEnumFieldOptions = `
syntax = "proto3";

package test;

enum TestEnumOption {
  reserved 2;
  option allow_alias = true;
  FIRST = 0;
  SECOND = 1;
  SEGUNDO = 1 [(my_enum_value_option) = 321];
}
`

const protoWithSingleQuoteReservedNames = `
syntax = "proto3";

package test;

message Channel {
  reserved 'thing', 'another';
  reserved "more", 'mixed';
  int64 id = 1;
  string name = 2;
  string description = 3;
}
`

var gpfPath = filepath.Join("testdata", "getProtoFiles")

func TestParseSingleQuoteReservedNames(t *testing.T) {
	r := strings.NewReader(protoWithSingleQuoteReservedNames)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Len(t, entry.Messages[0].ReservedNames, 4)
	assert.EqualValues(t,
		[]string{"thing", "another", "more", "mixed"},
		entry.Messages[0].ReservedNames,
	)
}

func TestParseIncludingImports(t *testing.T) {
	r := strings.NewReader(protoWithImports)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Equal(t, "testdata/test.proto", entry.Imports[0].Path)
}

func TestParseIncludingMessageOptions(t *testing.T) {
	r := strings.NewReader(protoWithMessageOptions)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Equal(t, "(ext.persisted)", entry.Messages[0].Options[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Options[0].Value)
}

func TestParseIncludingNestedMessageOptions(t *testing.T) {
	r := strings.NewReader(protoWithNestedMessageOptions)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Equal(t, "(ext.persisted)", entry.Messages[0].Options[0].Name)
	assert.Empty(t, entry.Messages[0].Options[0].Value)
	assert.Len(t, entry.Messages[0].Options[0].Aggregated, 2)
	assert.Equal(t, "opt1", entry.Messages[0].Options[0].Aggregated[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Options[0].Aggregated[0].Value)
	assert.Equal(t, "opt2", entry.Messages[0].Options[0].Aggregated[1].Name)
	assert.Equal(t, "false", entry.Messages[0].Options[0].Aggregated[1].Value)
}

func TestParseIncludingFieldOptions(t *testing.T) {
	r := strings.NewReader(protoWithFieldOptions)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Equal(t, "(personal)", entry.Messages[0].Fields[1].Options[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Fields[1].Options[0].Value)
	assert.Equal(t, "(owner)", entry.Messages[0].Fields[1].Options[1].Name)
	assert.Equal(t, "test", entry.Messages[0].Fields[1].Options[1].Value)
	assert.Len(t, entry.Messages[0].Maps, 1)
	assert.Equal(t, "string", entry.Messages[0].Maps[0].KeyType)
	assert.Equal(t, "attributes", entry.Messages[0].Maps[0].Field.Name)
	assert.Len(t, entry.Messages[0].Maps[0].Field.Options, 1)
	assert.Equal(t, "(personal)", entry.Messages[0].Maps[0].Field.Options[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Maps[0].Field.Options[0].Value)
}

func TestParseIncludingNestedFieldOptions(t *testing.T) {
	r := strings.NewReader(protoWithNestedFieldOptions)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Len(t, entry.Messages[0].Fields[3].Options, 2)
	assert.Equal(t, "(custom_options).personal", entry.Messages[0].Fields[3].Options[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Fields[3].Options[0].Value)
	assert.Equal(t, "(custom_options).internal", entry.Messages[0].Fields[3].Options[1].Name)
	assert.Equal(t, "false", entry.Messages[0].Fields[3].Options[1].Value)
}

func TestParseIncludingNestedFieldOptionsAggregated(t *testing.T) {
	r := strings.NewReader(protoWithNestedFieldOptionsAggregated)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Len(t, entry.Messages[0].Fields[2].Options, 1)
	assert.Equal(t, "(custom_options_commas)", entry.Messages[0].Fields[2].Options[0].Name)
	assert.Equal(t, "personal", entry.Messages[0].Fields[2].Options[0].Aggregated[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Fields[2].Options[0].Aggregated[0].Value)
	assert.Equal(t, "internal", entry.Messages[0].Fields[2].Options[0].Aggregated[1].Name)
	assert.Equal(t, "false", entry.Messages[0].Fields[2].Options[0].Aggregated[1].Value)
	assert.Equal(t, "owner", entry.Messages[0].Fields[2].Options[0].Aggregated[2].Name)
	assert.Equal(t, "some owner", entry.Messages[0].Fields[2].Options[0].Aggregated[2].Value)
	assert.Len(t, entry.Messages[0].Fields[3].Options, 1)
	assert.Equal(t, "(custom_options)", entry.Messages[0].Fields[3].Options[0].Name)
	assert.Equal(t, "personal", entry.Messages[0].Fields[3].Options[0].Aggregated[0].Name)
	assert.Equal(t, "true", entry.Messages[0].Fields[3].Options[0].Aggregated[0].Value)
	assert.Equal(t, "internal", entry.Messages[0].Fields[3].Options[0].Aggregated[1].Name)
	assert.Equal(t, "false", entry.Messages[0].Fields[3].Options[0].Aggregated[1].Value)
	assert.Equal(t, "owner", entry.Messages[0].Fields[3].Options[0].Aggregated[2].Name)
	assert.Equal(t, "some owner", entry.Messages[0].Fields[3].Options[0].Aggregated[2].Value)
}

func TestParseIncludingEnumFieldOptions(t *testing.T) {
	r := strings.NewReader(protoWithEnumFieldOptions)

	entry, err := Parse(r)
	assert.NoError(t, err)

	assert.Len(t, entry.Enums, 1)
	assert.Equal(t, "TestEnumOption", entry.Enums[0].Name)
	assert.Len(t, entry.Enums[0].EnumFields, 3)
	assert.Equal(t, "FIRST", entry.Enums[0].EnumFields[0].Name)
	assert.Equal(t, "SECOND", entry.Enums[0].EnumFields[1].Name)
	assert.Equal(t, "SEGUNDO", entry.Enums[0].EnumFields[2].Name)
	assert.Len(t, entry.Enums[0].EnumFields[2].Options, 1)
	assert.Equal(t, "(my_enum_value_option)", entry.Enums[0].EnumFields[2].Options[0].Name)
	assert.Equal(t, "321", entry.Enums[0].EnumFields[2].Options[0].Value)
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
