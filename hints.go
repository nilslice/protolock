package protolock

import (
	"errors"
	"strings"

	"github.com/emicklei/proto"
)

const (
	// CommentSkip tells the parse step to skip the comparable entity.
	CommentSkip = "@protolock:skip"
)

var (
	// ErrSkipEntry indicates that the CommentSkip hint was found.
	ErrSkipEntry = errors.New("protolock: skip entry comment encountered")
)

func checkComments(v interface{}) error {
	switch v.(type) {
	case *proto.Message:
		m := v.(*proto.Message)
		return hint(m.Comment)

	case *proto.Service:
		s := v.(*proto.Service)
		return hint(s.Comment)
	}

	return nil
}

func hint(c *proto.Comment) error {
	if c == nil {
		return nil
	}

	for _, line := range c.Lines {
		if strings.Contains(line, CommentSkip) {
			return ErrSkipEntry
		}
	}

	return nil
}
