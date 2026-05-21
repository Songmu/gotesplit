package examples

import (
	"fmt"
	"testing"
)

func TestRegular(t *testing.T) {}

func Example_withOutput() {
	fmt.Println(Hello())
	// Output: hi
}

func Example_noOutput() {
	// no output comment, go test won't run this
	_ = Hello()
}

func Example_unordered() {
	fmt.Println("a")
	fmt.Println("b")
	// Unordered output:
	// b
	// a
}
