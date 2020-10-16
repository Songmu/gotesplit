package gotesplit

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type genRegexp struct {
}

func (g *genRegexp) run(ctx context.Context, argv []string, outStream io.Writer, errStream io.Writer) error {
	if len(argv) < 3 {
		return fmt.Errorf("not enough arguments")
	}
	pkgs := argv[2:]
	total, err := strconv.Atoi(argv[0])
	if err != nil {
		return fmt.Errorf("invalid total: %s", err)
	}
	idx, err := strconv.Atoi(argv[1])
	if err != nil {
		return fmt.Errorf("invalid index: %s", err)
	}

	str, err := getOut(pkgs, total, idx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(outStream, str)
	return err
}
