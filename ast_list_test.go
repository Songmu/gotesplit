package gotesplit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"testing"
)

// chdirT changes cwd and restores it after the test.
func chdirT(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v (cwd=%s)", dir, err, orig)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})
}

func TestIsTestFunc(t *testing.T) {
	testCases := []struct {
		src    string
		expect bool
		desc   string
	}{
		{`package x; func TestFoo(t *testing.T) {}`, true, "regular Test"},
		{`package x; func TestMain(t *testing.T) {}`, true, "TestMain taking *T is a regular test"},
		{`package x; func TestX(t *foo.T) {}`, true, "renamed-import equivalent"},
		{`package x; func TestMain(m *testing.M) {}`, false, "TestMain *M is entry point (drop)"},
		{`package x; func BenchmarkFoo(b *testing.B) {}`, false, "Benchmark drops"},
		{`package x; func FuzzFoo(f *testing.F) {}`, false, "Fuzz drops"},
		{`package x; func Test() {}`, false, "prefix only"},
		{`package x; func Testxxx(t *testing.T) {}`, false, "lowercase suffix"},
		{`package x; func Bench() {}`, false, "not a test candidate"},
		{`package x; func Helper() {}`, false, "not a test candidate"},
		{`package x; func BenchmarkBad(s string) {}`, false, "invalid Benchmark silently drops"},
		{`package x; func TestBad(s string) {}`, false, "invalid Test silently drops"},
		{`package x; func TestExtra(t *testing.T, n int) {}`, false, "two parameters drops"},
		{`package x; func TestMain(a, b *testing.M) {}`, false, "multi-name field drops"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fn := parseFuncDecl(t, tc.src)
			got := isTestFunc(fn)
			if got != tc.expect {
				t.Errorf("expect: %v, got: %v", tc.expect, got)
			}
		})
	}
}

func TestASTvsGotestEquivalence(t *testing.T) {
	chdirT(t, "testdata/ast_fixture")

	testCases := []struct {
		pkgs []string
		tags string
		race bool
		desc string
	}{
		{[]string{"./simple"}, "", false, "simple"},
		{[]string{"./examples"}, "", false, "examples"},
		{[]string{"./withtags"}, "", false, "withtags-notag"},
		{[]string{"./withtags"}, "-tags=a", true, "withtags-taga with -race"},
		{[]string{"./racetag"}, "", false, "racetag-norace"},
		{[]string{"./racetag"}, "", true, "racetag-race"},
		{[]string{"./..."}, "", false, "multipkg"},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			expect, err := getTestListsFromPkgs(tc.pkgs, tc.tags, tc.race)
			if err != nil {
				t.Fatalf("gotest: unexpected error: %v", err)
			}
			got, err := getTestListsFromPkgsAST(tc.pkgs, tc.tags, tc.race)
			if err != nil {
				t.Fatalf("ast: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(expect, got) {
				t.Errorf("expect: %#v\ngot: %#v", expect, got)
			}
		})
	}
}

func TestGetTestListsFromPkgsASTPackageError(t *testing.T) {
	_, err := getTestListsFromPkgsAST([]string{"github.com/nonexistent/pkg-that-does-not-exist"}, "", false)
	if err == nil {
		t.Fatal("expect error for nonexistent package, got nil")
	}
}

func parseFuncDecl(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			return fn
		}
	}
	t.Fatal("no FuncDecl")
	return nil
}
