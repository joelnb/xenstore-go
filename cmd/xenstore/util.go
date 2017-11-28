package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// PrintFullWidth prints two values at the full width of the current terminal. Spacing is added so
// that the entire width is used
func PrintFullWidth(s1, s2 string) {
	w, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		w = 80
	}

	pLen := w - len(s1) - len(s2)
	if pLen < 5 {
		pLen = 5
	}

	fmt.Println(s1 + strings.Repeat(" ", pLen) + s2)
}

// Some methods may behave slightly differently if outputting to pipe (for easier grepping etc.)
func StdOutIsPipe() (bool, error) {
	if runtime.GOOS == "windows" {
		return true, nil
	}

	fi, err := os.Stdout.Stat()
	if err != nil {
		return true, err
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return true, nil
	}

	return false, nil
}
