package gotesplit

import (
	"context"
	"io"
)

var dispatch = map[string]runner{
	"regexp": &genRegexp{},
}

type runner interface {
	run(context.Context, []string, io.Writer, io.Writer) error
}
