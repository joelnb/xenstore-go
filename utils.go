package xenstore

import (
	"io/ioutil"
	"sync"
)

type xenStoreOperation uint32

const (
	XsDebug xenStoreOperation = iota
	XsDirectory
	XsRead
	XsGetPermissions
	XsWatch
	XsUnWatch
	XsStartTransaction
	XsEndTransaction
	XsIntroduce
	XsRelease
	XsGetDomainPath
	XsWrite
	XsMkdir
	XsRm
	XsSetPermissions
	XsWatchEvent
	XsError
	XsIsDomainIntroduced
	XsResume
	XsSetTarget
	XsRestrict
	XsResetWatches

	XsInvalid xenStoreOperation = 0xffff

	// XenStorePathSeparator is the separator between paths in XenStore. Parts of any path sent
	// to/received from XenStore should be joined with exactly 1 instance of this string. This is
	// not platform dependent.
	XenStorePathSeparator = "/"
)

var (
	requestCounter uint32 = 0x0
	counterMutex   *sync.Mutex

	NUL byte = 0x0
)

func init() {
	// Create the mutex used to synchronise access to the request counter variable.
	counterMutex = &sync.Mutex{}
}

// Event implements a XenStore event
type Event struct {
	Path  string
	Token string
}

// RequestID returns the next unique (for this session) request ID to use when contacting XenStore.
// RequestID synchronises access to the counter and is therefore safe to call across multiple
// goroutines.
func RequestID() uint32 {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	requestCounter += 1

	// Take 1 so that this begins at 0
	return requestCounter - 1
}

// ControlDomain checks whether the current Xen domain has the 'control_d' capability (will be true
// on Domain-0).
func ControlDomain() bool {
	r, err := ioutil.ReadFile("/proc/xen/capabilities")
	if err != nil {
		return false
	}

	if string(r) == "control_d\n" {
		return true
	}

	return false
}
