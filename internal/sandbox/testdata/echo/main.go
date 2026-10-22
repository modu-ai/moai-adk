// A static echo fixture runs without mounting host programs or shared libraries.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if _, err := fmt.Println(strings.Join(os.Args[1:], " ")); err != nil {
		os.Exit(1)
	}
}
