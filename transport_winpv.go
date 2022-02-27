//go:build windows
// +build windows

package xenstore

import (
	"fmt"
	"sync"
	// "time"

	// "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yusufpapurcu/wmi"
)

var lock sync.Mutex

type XenProjectXenStoreBase struct {
	Active       bool
	InstanceName string
	XenTime      uint64
	// XenTime      time.Time
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
	swb, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		return err
	}
	wmi.DefaultClient.SWbemServicesClient = swb
	defer wmi.DefaultClient.SWbemServicesClient.Close()

	var baseList []XenProjectXenStoreBase
	query := wmi.CreateQuery(&baseList, "")
	fmt.Println(query)

	baseDispatchList, err := wmi.QueryNamespaceRaw(query, &baseList, "root\\wmi")
	if err != nil {
		return err
	}
	defer func() {
		for _, item := range baseDispatchList {
			item.Release()
		}
	}()

	for i := range baseDispatchList {
		item := baseDispatchList[i]
		fmt.Printf("%+v\n", item)
		base := baseList[i]
		fmt.Printf("%+v\n", base)

		methodName := "AddSession"
		resultRaw, err := oleutil.CallMethod(item, methodName, "JoelSession")
		if err != nil {
			return fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
		}
		fmt.Printf("%+v\n", resultRaw)
	}

	return nil
}
