package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := getCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
