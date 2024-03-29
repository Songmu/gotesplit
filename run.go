package gotesplit

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jstemmer/go-junit-report/v2/gtr"
	"github.com/jstemmer/go-junit-report/v2/junit"
	parser "github.com/jstemmer/go-junit-report/v2/parser/gotest"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, total, idx uint, junitDir string, argv []string, outStream io.Writer, errStream io.Writer) error {
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
	if junitDir != "" {
		verbose := false
		for _, opt := range testOpts {
			if strings.HasSuffix(opt, "-v") {
				verbose = true
			}
			if strings.HasSuffix(opt, "-json") {
				return fmt.Errorf("-json output and -junitDir cannot be specified at the same time")
			}
		}
		if !verbose {
			testOpts = append([]string{"-v"}, testOpts...)
		}
	}

	testLists, err := getTestListsFromPkgs(pkgs, detectTags(testOpts), detectRace(testOpts))
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

	if junitDir != "" {
		if err := os.MkdirAll(junitDir, 0755); err != nil {
			return err
		}
	}

	var testArgsList [][]string

	if len(allPkgs) > 0 {
		args := append([]string{"test"}, testOpts...)
		args = append(args, allPkgs...)
		testArgsList = append(testArgsList, args)
	}
	for _, tl := range targetTests {
		args := append([]string{"test"}, testOpts...)
		run := "^(?:" + strings.Join(tl.list, "|") + ")$"
		args = append(args, "-run", run, tl.pkg)
		testArgsList = append(testArgsList, args)
	}

	for i, args := range testArgsList {
		report := goTest(args, outStream, errStream, junitDir)
		if err2 := report.err; err2 != nil {
			err = err2
		}
		if report.report != nil {
			if report.report.err != nil {
				log.Printf("failed to collect test report: %s\n", err)
			} else {
				fpath := filepath.Join(junitDir, fmt.Sprintf("junit-%d-%d.xml", idx, i))
				f, err := os.Create(fpath)
				if err != nil {
					log.Printf("failed to open file to store test report: %s\n", err)
				} else {
					defer f.Close()
					if err := writeJUnitReportXML(f, report.report.report); err != nil {
						log.Printf("failed to store test report: %s\n", err)
					}
				}
			}
		}
	}
	return err
}

type junitReport struct {
	report gtr.Report
	err    error
}

type testReport struct {
	err    error
	report *junitReport
}

func goTest(args []string, stdout, stderr io.Writer, junitDir string) *testReport {
	cmd := exec.Command("go", args...)
	log.Printf("running following go test:\n%s", cmd.String())

	var (
		outCloser io.Closer
		errCloser io.Closer
		outReader io.Reader
		errReader io.Reader
		outBuf    = bytes.NewBuffer(nil)
		eg        = &errgroup.Group{}
	)
	if junitDir == "" {
		cmd.Stdout = stdout
		cmd.Stderr = stderr
	} else {
		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return &testReport{
				err: err,
			}
		}
		defer outPipe.Close()
		outCloser = outPipe
		outReader = io.TeeReader(outPipe, outBuf)

		errPipe, err := cmd.StderrPipe()
		if err != nil {
			return &testReport{
				err: err,
			}
		}
		defer errPipe.Close()
		errCloser = errPipe
		errReader = io.TeeReader(errPipe, outBuf)
	}
	if err := cmd.Start(); err != nil {
		return &testReport{
			err: err,
		}
	}

	if junitDir != "" {
		eg.Go(func() error {
			defer outCloser.Close()
			_, err := io.Copy(stdout, outReader)
			if err != nil && errors.Is(err, os.ErrClosed) {
				err = nil
			}
			return err
		})
		eg.Go(func() error {
			defer errCloser.Close()
			_, err := io.Copy(stderr, errReader)
			if err != nil && errors.Is(err, os.ErrClosed) {
				err = nil
			}
			return err
		})
	}
	eg.Go(cmd.Wait)

	err := eg.Wait()
	ret := &testReport{
		err: err,
	}
	if junitDir != "" {
		report, err := parser.NewParser().Parse(outBuf)
		ret.report = &junitReport{
			report: report,
			err:    err,
		}
	}
	return ret
}

// Copied and pasted from
// https://github.com/jstemmer/go-junit-report/blob/v2.0.0/internal/gojunitreport/go-junit-report.go#L73-L93.
// So the license of the following line follows the original one.
func writeJUnitReportXML(w io.Writer, report gtr.Report) error {
	testsuites := junit.CreateFromReport(report, "")
	if _, err := fmt.Fprintf(w, xml.Header); err != nil {
		return err
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if err := enc.Encode(testsuites); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n")
	return err
}
