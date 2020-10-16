package gotesplit

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
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

	argv = fs.Args()
	if len(argv) < 1 {
		return errors.New("no subcommand specified")
	}
	rnr, ok := dispatch[argv[0]]
	if !ok {
		return fmt.Errorf("unknown subcommand or option: %s", argv[0])
	}
	return rnr.run(ctx, argv[1:], outStream, errStream)
}

func getTestListsFromPkgs(pkgs []string) ([]testList, error) {
	args := append([]string{"test", "-list", "."}, pkgs...)
	buf := &bytes.Buffer{}
	c := exec.Command("go", args...)
	c.Stdout = buf
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return nil, err
	}
	return getTestLists(buf.String()), nil
}

type testList struct {
	pkg  string
	list []string
}

func getTestLists(out string) []testList {
	var lists []testList
	var list []string
	for _, v := range strings.Split(out, "\n") {
		if strings.HasPrefix(v, "Test") {
			list = append(list, v)
			continue
		}
		if strings.HasPrefix(v, "ok ") {
			stuff := strings.Fields(v)
			if len(stuff) != 3 {
				continue
			}
			sort.Strings(list)
			lists = append(lists, testList{
				pkg:  stuff[1],
				list: list,
			})
			list = nil
		}
	}
	sort.Slice(lists, func(i, j int) bool {
		return len(lists[i].list) < len(lists[j].list)
	})
	return lists
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}
