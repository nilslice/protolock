package protolock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSPathToProtoPath(t *testing.T) {
	path := protopath("testdata/test.proto")
	p := protoPath(path)
	assert.Equal(t, "testdata:/:test.proto", string(p))
	assert.Equal(t, protopath("testdata:/:test.proto"), p)
}

func TestProtoPathToOSPath(t *testing.T) {
	path := protopath("testdata:/:test.proto")
	p := osPath(path)
	assert.Equal(t, protopath("testdata/test.proto"), p)
	assert.Equal(t, "testdata/test.proto", string(p))
}
