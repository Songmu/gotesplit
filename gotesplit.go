package gotesplit

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const cmdName = "gotesplit"

// Run the gotesplit
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Usage of %s:

  $ %s $pkg $total $index
  ^(?:TestAA|TestBB)$
  $ go test $pkg -run $(%[2]s $pkg $total $index)

Options:
`, fs.Name(), cmdName)
		fs.PrintDefaults()
	}
	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}

	pkg := argv[0]
	total, err := strconv.Atoi(argv[1])
	if err != nil {
		return fmt.Errorf("invalid total: %s", err)
	}
	idx, err := strconv.Atoi(argv[2])
	if err != nil {
		return fmt.Errorf("invalid index: %s", err)
	}

	str, err := getOut(pkg, total, idx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(outStream, str)
	return err
}

func getOut(pkg string, total, idx int) (string, error) {
	if total < 1 {
		return "", fmt.Errorf("invalid total: %d", total)
	}
	if idx >= total {
		return "", fmt.Errorf("index shoud be between 0 to total-1, but: %d (total:%d)", idx, total)
	}

	buf := &bytes.Buffer{}
	c := exec.Command("go", "test", "-list", ".", pkg)
	c.Stdout = buf
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", err
	}
	var list []string
	for _, v := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(v, "Test") {
			list = append(list, v)
		}
	}
	sort.Strings(list)

	testNum := len(list)
	minMemberPerGroup := testNum / total
	mod := testNum % total
	getOffset := func(i int) int {
		return minMemberPerGroup*i + int(math.Min(float64(i), float64(mod)))
	}
	from := getOffset(idx)
	to := getOffset(idx + 1)
	s := list[from:to]

	if len(s) == 0 {
		return "0^", nil
	}
	return "^(?:" + strings.Join(list[from:to], "|") + ")$", nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}
