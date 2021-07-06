package splittestgen

import (
	"sort"
	"strings"
)

type Packages []Package

type Package struct {
	Name  string
	Tests []string
}

type Tests []Test

type Test struct {
	Pkg  string
	Name string
}

func (p Packages) Tests() Tests {
	var tests []Test
	for _, pkg := range p {
		for _, testName := range pkg.Tests {
			tests = append(tests, Test{Pkg: pkg.Name, Name: testName})
		}
	}
	return tests
}

func (t Tests) DevideEquallyBy(parallel int) []Tests {
	div := len(t) / parallel
	mod := len(t) % parallel
	var divided []Tests
	for i := 0; i < parallel; i++ {
		start := i * div
		end := (i + 1) * div
		if i <= mod {
			start += i
			end += i + 1
		}
		divided = append(divided, t[start:end])
	}
	return divided
}

type Command struct {
	Pkg   string
	Tests []string
}

func (t Tests) Commands() []Command {
	var commands []Command
	var c Command
	for _, test := range t {
		if c.Pkg == "" {
			c.Pkg = test.Pkg
		} else if test.Pkg != c.Pkg {
			commands = append(commands, c)
			c = Command{Pkg: test.Pkg}
		}

		c.Tests = append(c.Tests, test.Name)
	}
	if c.Pkg != "" {
		commands = append(commands, c)
	}
	return commands
}

func (c Command) Args() []string {
	return []string{
		c.Pkg,
		"-run",
		"^(?:" + strings.Join(c.Tests, "|") + ")$",
	}
}

func GetPackages(out string) Packages {
	var packages Packages
	var list []string
	for _, v := range strings.Split(out, "\n") {
		if strings.HasPrefix(v, "Test") || strings.HasPrefix(v, "Example") {
			list = append(list, v)
			continue
		}
		if strings.HasPrefix(v, "ok ") {
			stuff := strings.Fields(v)
			if len(stuff) != 3 {
				continue
			}
			sort.Strings(list)
			packages = append(packages, Package{
				Name:  stuff[1],
				Tests: list,
			})
			list = nil
		}
	}
	sort.Slice(packages, func(i, j int) bool {
		cmp := len(packages[i].Tests) - len(packages[j].Tests)
		if cmp != 0 {
			return cmp < 0
		}
		return strings.Compare(packages[i].Name, packages[j].Name) < 0
	})
	return packages
}
