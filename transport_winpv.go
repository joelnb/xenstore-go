//go:build windows
// +build windows

package xenstore

import (
	"fmt"
	"sync"
	// "time"

	"github.com/go-ole/go-ole"
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

	fmt.Println("Asking for base list")
	baseDispatchList, err := wmi.QueryNamespaceRaw(query, &baseList, "root\\wmi")
	if err != nil {
		return err
	}
	defer func() {
		for _, item := range baseDispatchList {
			item.Release()
		}
	}()
	fmt.Println("Got base list")

	var sessionId int32
	for i := range baseDispatchList {
		item := baseDispatchList[i]
		fmt.Printf("%+v\n", item)
		base := baseList[i]
		fmt.Printf("%+v\n", base)

		fmt.Println("Calling AddSession")
		methodName := "AddSession"

		sessionResultRaw := new(ole.VARIANT)
		ole.VariantInit(sessionResultRaw)

		resultRaw, err := oleutil.CallMethod(item, methodName, "JoelSession", sessionResultRaw)
		if err != nil {
			return fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
		}
		defer resultRaw.Clear()

		fmt.Println("Called AddSession")

		// Need the ID of the created session
		fmt.Printf("%+v\n", resultRaw)
		// result := resultRaw.ToIDispatch()
		// defer result.Release()
		fmt.Printf("%+v\n", sessionResultRaw)
		sessionId = sessionResultRaw.Value().(int32)
		fmt.Printf("%+v\n", sessionId)
	}

	fmt.Println("Asking for session list")
	var sessionList []XenProjectXenStoreSession
	sessionQuery := fmt.Sprintf("SELECT Active, ID, InstanceName, SessionID FROM XenProjectXenStoreSession WHERE SessionId=%d", sessionId)
	sessionDispatchList, err := wmi.QueryNamespaceRaw(sessionQuery, &sessionList, "root\\wmi")
	if err != nil {
		return err
	}
	defer func() {
		for _, item := range sessionDispatchList {
			item.Release()
		}
	}()

	fmt.Println("Got session list")
	fmt.Printf("%+v\n", sessionDispatchList)

	for i := range sessionDispatchList {
		// session := sessionList[i]
		sessionDispatch := sessionDispatchList[i]

		methodName := "GetValue"
		resultRaw, err := oleutil.CallMethod(sessionDispatch, methodName, "vm")
		if err != nil {
			return fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
		}
		defer resultRaw.Clear()

		fmt.Printf("%+v\n", resultRaw)
		result := resultRaw.ToString()
		fmt.Printf("%+v\n", result)
	}

	return nil
}
