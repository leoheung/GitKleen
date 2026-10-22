// GitKleen-helper runs git for GitKleen on another computer, which starts it
// there over SSH.
//
//	gitkleen-helper serve   take GitKleen's calls on standard input
package main

import (
	"fmt"
	"os"

	"gitkleen/internal/askpass"
	"gitkleen/internal/helper"
)

func main() {
	// Run by git for a prompt, hand it to the helper, which asks GitKleen.
	if code, ok := askpass.Main(os.Args); ok {
		os.Exit(code)
	}
	if len(os.Args) != 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: gitkleen-helper serve")
		os.Exit(2)
	}
	if err := helper.Serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "gitkleen-helper:", err)
		os.Exit(1)
	}
}
