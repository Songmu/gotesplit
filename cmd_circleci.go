package gotesplit

import (
	"context"
	"io"
	"os"
)

type cmdCircleCI struct {
}

func (c *cmdCircleCI) run(ctx context.Context, argv []string, outStream io.Writer, errStream io.Writer) error {
	argv = append([]string{os.Getenv("CIRCLE_NODE_TOTAL"), os.Getenv("CIRCLE_NODE_INDEX")}, argv...)
	return (&cmdRun{}).run(ctx, argv, outStream, errStream)
}
