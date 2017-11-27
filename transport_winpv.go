//go:build windows
// +build windows

package xenstore

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yusufpapurcu/wmi"
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

func oleInt64(item *ole.IDispatch, prop string) (int64, error) {
	v, err := oleutil.GetProperty(item, prop)
	if err != nil {
		return 0, err
	}
	defer v.Clear()

	i := int64(v.Val)
	return i, nil
}

func initialiseWmi() (bool, error) {
	lock.Lock()
	runtime.LockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		oleerr := err.(*ole.OleError)

		if oleerr.Code() != ole.S_OK && oleerr.Code() != 0x00000001 {
			return false, err
		}

		return false, nil
	}
	return true, nil
}

func uninitialiseWmi(was bool) {
	if was {
		ole.CoUninitialize()
	}

	lock.Unlock()
	runtime.UnlockOSThread()
}

func internalOleWmiQuery(query string, connectServerArgs ...interface{}) (*ole.IDispatch, error) {
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return nil, err
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", connectServerArgs...)
	if err != nil {
		return nil, err
	}
	service := serviceRaw.ToIDispatch()
	defer serviceRaw.Clear()

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return nil, err
	}
	result := resultRaw.ToIDispatch()
	defer resultRaw.Clear()

	return result, nil
}

func internalOleWmiQuerySingle(query string, connectServerArgs ...interface{}) (*ole.IDispatch, error) {
	result, err := internalOleWmiQuery(query, connectServerArgs...)
	if err != nil {
		return nil, err
	}

	count, err := oleInt64(result, "Count")
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errors.New("expected a single valued WMI result")
	}

	itemRaw, err := oleutil.CallMethod(result, "ItemIndex", 0)
	if err != nil {
		return nil, err
	}
	item := itemRaw.ToIDispatch()
	defer itemRaw.Clear()

	return item, nil
}

func NewWinPVTransport() error {
	var base []XenProjectXenStoreBase
	query := wmi.CreateQuery(&base, "")
	fmt.Println(query)

	result, err := OleWMIQuerySingle(query, nil, "root\\wmi")

	// count, err := oleInt64(result, "Count")
	// if err != nil {
	//  log.Fatal(err)
	// }
	// fmt.Println(count)

	// Initialize a slice with Count capacity
	// dv.Set(reflect.MakeSlice(dv.Type(), 0, int(count)))

	// Call AddSession method on the XenProjectXenStoreBase WMI object
	mresRaw, err := oleutil.CallMethod(result, "AddSession")
	if err != nil {
		return err
	}
	mres := mresRaw.ToIDispatch()
	defer mresRaw.Clear()

	fmt.Printf("%+v\n", mres)

	return nil
}

func OleWMIQuery(query string, connectServerArgs ...interface{}) (*ole.IDispatch, error) {
	was, err := initialiseWmi()
	if err != nil {
		return nil, err
	}
	defer uninitialiseWmi(was)

	return internalOleWmiQuery(query, connectServerArgs...)
}

func OleWMIQuerySingle(query string, connectServerArgs ...interface{}) (*ole.IDispatch, error) {
	was, err := initialiseWmi()
	if err != nil {
		return nil, err
	}
	defer uninitialiseWmi(was)

	return internalOleWmiQuerySingle(query, connectServerArgs...)
}
