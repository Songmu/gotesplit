package gotesplit

import (
	"os"
	"reflect"
	"testing"
)

func TestGetTestList(t *testing.T) {
	const sample = `TestDoCreate
TestCommandGet
TestLook
TestDoGet_bulk
TestCommandList
TestCommandListUnique
TestCommandListUnknown
TestDoList_query
TestDoList_unique
TestDoList_unknownRoot
TestDoList_notPermittedRoot
TestDoList_withSystemHiddenDir
TestDoRoot
TestDetectLocalRepoRoot
TestDetectVCSAndRepoURL
TestLocalRepositoryFromFullPath
TestNewLocalRepository
TestLocalRepositoryRoots
TestList_Symlink
TestList_Symlink_In_Same_Directory
TestFindVCSBackend
TestLocalRepository_VCS
TestLocalRepositoryRoots_URLMatchLocalRepositoryRoots
TestNewRemoteRepository
TestNewRemoteRepository_vcs_error
TestNewRemoteRepository_error
TestNewURL
TestConvertGitURLHTTPToSSH
TestNewURL_err
TestFillUsernameToPath_err
TestVCSBackend
TestCvsDummyBackend
TestBranchOptionIgnoredErrors
ok      github.com/x-motemen/ghq        0.114s
TestRunInDirSilently
TestRun
TestRunInDir
TestRunSilently
ok      github.com/x-motemen/ghq/cmdutil        0.059s
?       github.com/x-motemen/ghq/hoge   [no test files]
TestLog
ok      github.com/x-motemen/ghq/logger 0.106s`

	var expect []testList = []testList{{
		pkg: "github.com/x-motemen/ghq/logger",
		list: []string{
			"TestLog",
		}}, {
		pkg: "github.com/x-motemen/ghq/cmdutil",
		list: []string{
			"TestRun",
			"TestRunInDir",
			"TestRunInDirSilently",
			"TestRunSilently",
		}}, {
		pkg: "github.com/x-motemen/ghq",
		list: []string{
			"TestBranchOptionIgnoredErrors",
			"TestCommandGet",
			"TestCommandList",
			"TestCommandListUnique",
			"TestCommandListUnknown",
			"TestConvertGitURLHTTPToSSH",
			"TestCvsDummyBackend",
			"TestDetectLocalRepoRoot",
			"TestDetectVCSAndRepoURL",
			"TestDoCreate",
			"TestDoGet_bulk",
			"TestDoList_notPermittedRoot",
			"TestDoList_query",
			"TestDoList_unique",
			"TestDoList_unknownRoot",
			"TestDoList_withSystemHiddenDir",
			"TestDoRoot",
			"TestFillUsernameToPath_err",
			"TestFindVCSBackend",
			"TestList_Symlink",
			"TestList_Symlink_In_Same_Directory",
			"TestLocalRepositoryFromFullPath",
			"TestLocalRepositoryRoots",
			"TestLocalRepositoryRoots_URLMatchLocalRepositoryRoots",
			"TestLocalRepository_VCS",
			"TestLook",
			"TestNewLocalRepository",
			"TestNewRemoteRepository",
			"TestNewRemoteRepository_error",
			"TestNewRemoteRepository_vcs_error",
			"TestNewURL",
			"TestNewURL_err",
			"TestVCSBackend",
		}}}
	got := getTestLists(sample)
	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expect: %#v\ngot: %#v", expect, got)
	}
}

func TestDetectTags(t *testing.T) {
	testCases := []struct {
		input  []string
		expect string
	}{
		{[]string{"aa", "bb"}, ""},
		{[]string{"aa", "-tags", "bb"}, "-tags=bb"},
		{[]string{"aa", "--tags=ccc", "bb"}, "--tags=ccc"},
		{[]string{"aa", "-tags"}, "-tags"},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			out := detectTags(tc.input)
			if out != tc.expect {
				t.Errorf("got: %s, expect: %s", out, tc.expect)
			}
		})
	}
}

func TestDetectRace(t *testing.T) {
	testCases := []struct {
		input  []string
		expect bool
		desc   string
	}{
		{[]string{"-race"}, true, "-race only"},
		{[]string{"-tags", "aaa", "-race", "-bench"}, true, "-race with other flags"},
		{[]string{"--race", "-p", "1"}, true, "--race with other flags"},
		{[]string{}, false, "no flags"},
		{[]string{"-short", "-p", "1"}, false, "flags without -race"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			out := detectRace(tc.input)
			if out != tc.expect {
				t.Errorf("got: %t, expect: %t", out, tc.expect)
			}
		})
	}
}

func TestGetTestListFromPkgs(t *testing.T) {
	if err := os.Chdir("testdata/withtags"); err != nil {
		wd, _ := os.Getwd()
		t.Fatalf("unexpected error: %v, cwd: %s", err, wd)
	}

	expect := []testList{{
		pkg: "github.com/Songmu/gotesplit/testdata/withtags",
		list: []string{
			"TestNoTag",
			"TestTagA",
		},
	}}

	got, err := getTestListsFromPkgs([]string{"."}, "-tags=a", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expect: %#v\ngot: %#v", expect, got)
	}
}
