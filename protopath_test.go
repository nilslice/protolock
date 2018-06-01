package protolock

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var osp = filepath.Join("testdata", "test.proto")

func TestOSPathToProtoPath(t *testing.T) {
	path := protopath(osp)
	p := protoPath(path)
	assert.Equal(t, "testdata:/:test.proto", string(p))
	assert.Equal(t, protopath("testdata:/:test.proto"), p)
}

func TestProtoPathToOSPath(t *testing.T) {
	path := protopath("testdata:/:test.proto")
	p := osPath(path)
	assert.Equal(t, protopath(osp), p)
	assert.Equal(t, osp, string(p))
}
