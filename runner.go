package gotesplit

import (
	"context"
	"io"
)

var dispatch = map[string]runner{
	"regexp": &cmdRegexp{},
	"run":    &cmdRun{},
}

type runner interface {
	run(context.Context, []string, io.Writer, io.Writer) error
}
