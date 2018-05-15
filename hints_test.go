package protolock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const hintSkip = `syntax = "proto3";
package dataset;

// @protolock:skip
message Channel {
  reserved 6, 8 to 11;
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  int32 age = 5;
}

message NextRequest {}
// this text before our hint shouln't matter +(#*)//.~  @protolock:skip
message PreviousRequest {}

// @protolock:skip
// @protolock:no-impl <- not a real hint, should pick up skip for ChannelChanger
service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

func TestHints(t *testing.T) {
	lock := parseTestProto(t, hintSkip)

	for _, def := range lock.Definitions {
		t.Run("skip:messages", func(t *testing.T) {
			assert.Len(t, def.Def.Messages, 1)
		})
		t.Run("skip:services", func(t *testing.T) {
			assert.Len(t, def.Def.Services, 0)
		})
	}
}
