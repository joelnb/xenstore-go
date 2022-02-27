package xenstore

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	// TODO: Match start/end of text not line
	pathRegex        = regexp.MustCompile(`^[a-zA-Z0-9-/_@]+\x00?$`)
	watchPathRegex   = regexp.MustCompile(`^@(?:introduceDomain|releaseDomain)\x00?$`)
	permissionsRegex = regexp.MustCompile(`^[wrbn]\d+$`)
)

// JoinXenStorePath concatenates parts of a path together with the XenStorePathSeparator,
// ensuring that exactly 1 instance of the XenStorePathSeparator is used.
func JoinXenStorePath(paths ...string) string {
	var fullpath string

	for i, path := range paths {
		firstElem := (i == 0)

		if strings.HasSuffix(path, XenStorePathSeparator) && (!firstElem || len(paths) > 1) {
			path = path[:len(path)-1]
		}

		if !firstElem {
			if strings.HasPrefix(path, XenStorePathSeparator) {
				path = path[1:]
			}

			fullpath = fullpath + XenStorePathSeparator + path
		} else {
			fullpath = fullpath + path
		}
	}

	return fullpath
}

// UnixSocketPath gets the current path to the XenStore unix socket on this system
func UnixSocketPath() string {
	if e := os.Getenv("XENSTORED_PATH"); e != "" {
		return e
	}

	rundir := os.Getenv("XENSTORED_RUNDIR")
	if rundir == "" {
		rundir = "/var/run/xenstored"
	}

	return path.Join(rundir, "socket")
}

// XenBusPath returns the path to the XenBus device on this system
func XenBusPath() string {
	f, err := os.Open("/dev/xen/xenbus")
	if err != nil && runtime.GOOS == "linux" {
		return "/proc/xen/xenbus"
	}
	f.Close()

	switch runtime.GOOS {
	case "netbsd":
		return "/kern/xen/xenbus"
	default:
		return "/dev/xen/xenbus"
	}
}

// ValidPath returns a bool representing whether the provided string is a valid
// XenStore path.
func ValidPath(path string) bool {
	// Paths longer than 3072 bytes are forbidden & absolute paths have a higher limit
	maxLen := 2048
	if filepath.IsAbs(path) {
		maxLen = 3072
	}

	// Disallow if too long or not matching the regex
	if len(path) > maxLen || !pathRegex.Match([]byte(path)) {
		return false
	}

	// Some more specific rules regarding separators
	if (len(path) > 1 && strings.HasSuffix(path, XenStorePathSeparator)) ||
		strings.Contains(path, XenStorePathSeparator+XenStorePathSeparator) {
		return false
	}
	return true
}

// ValidWatchPath returns a bool representing whether the provided string is a valid watch
// path - a special case of XenStore paths.
func ValidWatchPath(path string) bool {
	if watchPathRegex.Match([]byte(path)) {
		return true
	}
	return ValidPath(path)
}

// ValidPermissions checks if a set of permission specifications for validity & returns
// true only if all are valid.
func ValidPermissions(permissions ...string) bool {
	for _, perm := range permissions {
		if !permissionsRegex.Match([]byte(perm)) {
			return false
		}
	}
	return true
}
