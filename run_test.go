package splittestgen_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	const total = 3
	for _, c := range []struct {
		index    int
		expected string
	}{
		{
			index:    0,
			expected: "go test github.com/minoritea/go-splittestgen/testdata/group1 -run '^(?:Test1|Test2)$'",
		},
		{
			index:    1,
			expected: "go test github.com/minoritea/go-splittestgen/testdata/group1 -run '^(?:Test3|Test4)$'",
		},
		{
			index:    2,
			expected: "go test github.com/minoritea/go-splittestgen/testdata/group2 -run '^(?:TestCases)$'",
		},
	} {
		index, expected := c.index, c.expected
		t.Run(fmt.Sprintf("case-of-index-%d", index), func(t *testing.T) {
			t.Parallel()
			output, err := exec.Command(
				"sh", "-c", fmt.Sprintf(` \
				  (cd testdata && go test ./... -list .) | \
					go run ./cmd/go-splittestgen -index %d -total %d \
				`, index, total),
			).CombinedOutput()

			if err != nil {
				t.Fatal(err)
			}

			if strings.TrimSpace(string(output)) != expected {
				t.Log("The actual output is\n", string(output))
				t.Log("The expected output is\n", expected)
				t.Fatal("The command output is not equal to the expected value")
			}
		})
	}
}
