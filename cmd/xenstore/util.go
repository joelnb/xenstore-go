package main

import (
	"os"
	"runtime"
)

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
