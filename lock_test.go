package protolock

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const simpleProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
	rpc Next(stream NextRequest) returns (Channel);
	rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noUsingReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  reserved 4, 8 to 11;
  reserved "foo", "bar";  
  int64 id = 1;
  string name = 2;
  string description = 3;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
	rpc Next(stream NextRequest) returns (Channel);
	rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const usingReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

func TestParseOnReader(t *testing.T) {
	r := strings.NewReader(simpleProto)
	_, err := parse(r)
	assert.NoError(t, err)
}

func TestCompareWithKnownIssue(t *testing.T) {
	cur := strings.NewReader(noUsingReservedFieldsProto)
	upd := strings.NewReader(usingReservedFieldsProto)

	curEntry, err := parse(cur)
	assert.NoError(t, err)
	curLock := Protolock{
		Definitions: []Definition{
			{
				Filepath: "NA",
				Def:      curEntry,
			},
		},
	}

	updEntry, err := parse(upd)
	assert.NoError(t, err)
	updLock := Protolock{
		Definitions: []Definition{
			{
				Filepath: "NA",
				Def:      updEntry,
			},
		},
	}

	report, err := compare(curLock, updLock)
	assert.Error(t, err)
	assert.Len(t, report.Warnings, 3)
}

func toJSON(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return b
}
