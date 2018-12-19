// +build windows

package xenstore

import (
	"fmt"
	"sync"
	"time"

	"github.com/mholt/wmi"
)

var lock sync.Mutex

type XenProjectXenStoreBase struct {
	Active       bool
	InstanceName string
	XenTime      time.Time
}

type XenProjectXenStoreSession struct {
	Active       bool
	ID           string
	InstanceName string
	SessionID    uint32
}

type XenProjectXenStoreWatchEvent struct {
	EventID string
}

type XenProjectXenStoreUnsuspendedEvent struct {
	ID        string
	SessionID uint32
}

func NewWinPVTransport() error {
	result, err := wmi.CallMethod(nil, "XenProjectXenStoreBase", "AddSession", []interface{}{})
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	return nil
}
