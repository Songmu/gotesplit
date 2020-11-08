package gotesplit

import (
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strings"
)

func run(ctx context.Context, total, idx uint, junitfile string, argv []string, outStream io.Writer, errStream io.Writer) error {
	if idx >= total {
		return fmt.Errorf("`index` should be the range from 0 to `total`-1, but: %d (total:%d)", idx, total)
	}

	l := len(argv)
	var (
		pkgs     []string
		testOpts []string
	)
	for i := 0; i < l; i++ {
		pkg := argv[i]
		if pkg == "--" {
			testOpts = argv[i+1:]
			break
		}
		pkgs = append(pkgs, pkg)
	}
	if junitfile != "" {
		verbose := false
		for _, opt := range testOpts {
			if strings.HasSuffix(opt, "-v") {
				verbose = true
			}
			if strings.HasSuffix(opt, "-json") {
				return fmt.Errorf("-json output and -junitfile cannot be specified at the same time")
			}
		}
		if !verbose {
			testOpts = append([]string{"-v"}, testOpts...)
		}
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
	testNum := uint(len(testListStrs))
	minMemberPerGroup := testNum / total
	mod := testNum % total
	getOffset := func(i uint) uint {
		return minMemberPerGroup*i + uint(math.Min(float64(i), float64(mod)))
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
