package group2_test

import "testing"

func TestMain(m *testing.M) { m.Run() }
func TestCases(t *testing.T) {
	cases := []string{"A", "B"}
	for _, name := range cases {
		t.Run("case-"+name, func(t *testing.T) { t.Logf("Case %s is OK", name) })
	}
}
