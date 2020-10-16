package gotesplit

import (
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type cmdRun struct {
}

func (c *cmdRun) run(ctx context.Context, argv []string, outStream io.Writer, errStream io.Writer) error {
	if len(argv) < 3 {
		return fmt.Errorf("not enough arguments")
	}
	total, err := strconv.Atoi(argv[0])
	if err != nil {
		return fmt.Errorf("invalid total: %s", err)
	}
	idx, err := strconv.Atoi(argv[1])
	if err != nil {
		return fmt.Errorf("invalid index: %s", err)
	}
	if total < 1 {
		return fmt.Errorf("invalid total: %d", total)
	}
	if idx >= total {
		return fmt.Errorf("index shoud be between 0 to total-1, but: %d (total:%d)", idx, total)
	}

	l := len(argv)
	var (
		pkgs     []string
		testOpts []string
	)
	for i := 2; i < l; i++ {
		pkg := argv[i]
		if pkg == "--" {
			testOpts = argv[i+1:]
			break
		}
		pkgs = append(pkgs, pkg)
	}
	testLists, err := getTestListsFromPkgs(pkgs)
	if err != nil {
		return err
	}
	testListMap := make(map[string]testList, len(testLists))
	for _, tl := range testLists {
		testListMap[tl.pkg] = tl
	}
	const delim = ":::"
	var testListStrs []string
	for _, v := range testLists {
		for _, t := range v.list {
			testListStrs = append(testListStrs, v.pkg+delim+t)
		}
	}
	testNum := len(testListStrs)
	minMemberPerGroup := testNum / total
	mod := testNum % total
	getOffset := func(i int) int {
		return minMemberPerGroup*i + int(math.Min(float64(i), float64(mod)))
	}
	from := getOffset(idx)
	to := getOffset(idx + 1)
	var (
		targetTests []testList
		allPkgs     []string
		currList    testList
	)
	addList := func(l testList) {
		tl := testListMap[l.pkg]
		if len(tl.list) == len(l.list) {
			allPkgs = append(allPkgs, l.pkg)
		} else {
			targetTests = append(targetTests, l)
		}
	}
	for _, v := range testListStrs[from:to] {
		stuff := strings.Split(v, delim)
		pkg := stuff[0]
		test := stuff[1]
		if pkg != currList.pkg {
			if currList.pkg != "" {
				addList(currList)
			}
			currList = testList{pkg: pkg}
		}
		currList.list = append(currList.list, test)
	}
	if len(currList.list) > 0 {
		addList(currList)
	}

	err = nil
	if len(allPkgs) > 0 {
		args := append([]string{"test"}, testOpts...)
		args = append(args, allPkgs...)
		cmd := exec.Command("go", args...)
		fmt.Fprintln(errStream, cmd.String())
		cmd.Stderr = errStream
		cmd.Stdout = outStream
		if err2 := cmd.Run(); err2 != nil {
			err = err2
		}
	}
	for _, tl := range targetTests {
		args := append([]string{"test"}, testOpts...)
		run := "^(?:" + strings.Join(tl.list, "|") + ")$"
		args = append(args, "-run", run, tl.pkg)
		cmd := exec.Command("go", args...)
		fmt.Fprintln(errStream, cmd.String())
		cmd.Stderr = errStream
		cmd.Stdout = outStream
		if err2 := cmd.Run(); err2 != nil {
			err = err2
		}
	}
	return err
}
