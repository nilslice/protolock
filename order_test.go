package protolock

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ignoreArg = ""

func TestOrder(t *testing.T) {
	// verify that the re-production of the same Protolock encoded as json
	// is equivalent to any previously encoded version of the same Protolock
	f, err := os.Open("proto.lock")
	assert.NoError(t, err)

	current, err := protolockFromReader(f)
	assert.NoError(t, err)

	r, err := Commit(ignoreArg)
	assert.NoError(t, err)
	assert.NotNil(t, r)

	updated, err := protolockFromReader(r)
	assert.NoError(t, err)

	assert.Equal(t, current, updated)

	a, err := json.Marshal(current)
	assert.NoError(t, err)
	b, err := json.Marshal(updated)
	assert.NoError(t, err)

	assert.Equal(t, a, b)
}
