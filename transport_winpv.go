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

type WinPVTransport struct{}

func (t *WinPVTransport) Close() {
	wmi.DefaultClient.SWbemServicesClient.Close()
}

func (t *WinPVTransport) GetBase() (*XenProjectXenStoreBase, error) {
	fmt.Println("GetBase: Asking for base list")

	var baseList []XenProjectXenStoreBaseProps
	baseQuery := "SELECT Active, InstanceName, XenTime FROM XenProjectXenStoreBase"
	baseDispatchList, err := wmi.QueryNamespaceRaw(baseQuery, &baseList, "root\\wmi")
	if err != nil {
		return nil, err
	}

	// defer func() {
	// 	for _, item := range baseDispatchList {
	// 		item.Release()
	// 	}
	// }()

	fmt.Println("GetBase: Got base list")

	if len(baseList) != len(baseDispatchList) {
		return nil, fmt.Errorf("WinPVTransport.GetBase: unexpected length mismatch (baseList=%d, baseDispatchList=%d)",
			len(baseList),
			len(baseDispatchList))
	} else if len(baseList) == 0 {
		return nil, fmt.Errorf("WinPVTransport.GetBase: Got 0-length list")
	} else if len(baseList) > 1 {
		return nil, fmt.Errorf("WinPVTransport.GetBase: Unexpected multiple XenProjectXenStoreBase returned")
	}

	return &XenProjectXenStoreBase{
		disp:       baseDispatchList[0],
		Properties: baseList[0],
	}, nil
}

func (t *WinPVTransport) GetSession(sessionId int32) (*XenProjectXenStoreSession, error) {
	fmt.Println("GetSession: Asking for session list")

	var sessionList []XenProjectXenStoreSessionProps
	sessionQuery := fmt.Sprintf("SELECT Active, ID, InstanceName, SessionID FROM XenProjectXenStoreSession WHERE SessionId=%d", sessionId)
	sessionDispatchList, err := wmi.QueryNamespaceRaw(sessionQuery, &sessionList, "root\\wmi")
	if err != nil {
		return nil, err
	}

	// defer func() {
	// 	for _, item := range sessionDispatchList {
	// 		item.Release()
	// 	}
	// }()

	if len(sessionList) != len(sessionDispatchList) {
		return nil, fmt.Errorf("WinPVTransport.GetSession: unexpected length mismatch (sessionList=%d, sessionDispatchList=%d)",
			len(sessionList),
			len(sessionDispatchList))
	} else if len(sessionList) == 0 {
		return nil, fmt.Errorf("WinPVTransport.GetSession: Got 0-length list")
	} else if len(sessionList) > 1 {
		return nil, fmt.Errorf("WinPVTransport.GetSession: Unexpected multiple XenProjectXenStoreBase returned")
	}

	return &XenProjectXenStoreSession{
		disp:       sessionDispatchList[0],
		Properties: sessionList[0],
	}, nil
}

func (b *XenProjectXenStoreBase) AddSession(name string) (int32, error) {
	var sessionId int32

	fmt.Println("Calling AddSession")

	sessionResultRaw := new(ole.VARIANT)
	ole.VariantInit(sessionResultRaw)
	defer sessionResultRaw.Clear()

	methodName := "AddSession"
	resultRaw, err := oleutil.CallMethod(b.disp, methodName, name, sessionResultRaw)
	if err != nil {
		return 0, fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	fmt.Println("Called AddSession")

	// Need the ID of the created session
	fmt.Printf("%+v\n", resultRaw)
	// result := resultRaw.ToIDispatch()
	// defer result.Release()
	fmt.Printf("%+v\n", sessionResultRaw)

	sessionId = sessionResultRaw.Value().(int32)
	return sessionId, nil
}

func (s *XenProjectXenStoreSession) GetValue(path string) (string, error) {
	valueResultRaw := new(ole.VARIANT)
	ole.VariantInit(valueResultRaw)
	defer valueResultRaw.Clear()

	methodName := "GetValue"
	resultRaw, err := oleutil.CallMethod(s.disp, methodName, path, valueResultRaw)
	if err != nil {
		return "", fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	return valueResultRaw.ToString(), nil
}

type XenProjectXenStoreBase struct {
	disp       *ole.IDispatch
	Properties XenProjectXenStoreBaseProps
}

type XenProjectXenStoreBaseProps struct {
	Active       bool
	InstanceName string
	XenTime      uint64
}

type XenProjectXenStoreSession struct {
	disp       *ole.IDispatch
	Properties XenProjectXenStoreSessionProps
}

type XenProjectXenStoreSessionProps struct {
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

func getBase() (*XenProjectXenStoreBase, error) {
	return nil, nil
}

func initWmi() error {
	swb, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		return err
	}
	wmi.DefaultClient.SWbemServicesClient = swb
	return nil
}

func NewWinPVTransport() error {
	if err := initWmi(); err != nil {
		return err
	}

	transport := WinPVTransport{}
	defer transport.Close()

	base, err := transport.GetBase()
	if err != nil {
		return err
	}

	sessionId, err := base.AddSession("JoelSession")
	if err != nil {
		return err
	}

	session, err := transport.GetSession(sessionId)
	if err != nil {
		return err
	}

	value, err := session.GetValue("vm")
	if err != nil {
		return err
	}

	fmt.Println(value)

	return nil
}
