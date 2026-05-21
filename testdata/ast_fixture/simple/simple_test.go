package simple

import (
	"fmt"
	"os"
	"testing"
)

func TestA(t *testing.T) {}
func TestB(t *testing.T) {}
func TestMain(m *testing.M) { os.Exit(m.Run()) }
func ExampleHello() {
	fmt.Println(Hello())
	// Output: hello
}
