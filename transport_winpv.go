//go:build windows
// +build windows

package xenstore

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/yusufpapurcu/wmi"
)

var lock sync.Mutex

type WinPVTransport struct{}

func (t *WinPVTransport) Close() {
	wmi.DefaultClient.SWbemServicesClient.Close()
}

func (t *WinPVTransport) GetBase() (*XenProjectXenStoreBase, error) {
	log.Debug("GetBase: Requesting base list")

	var baseList []XenProjectXenStoreBaseProps
	baseQuery := "SELECT Active, InstanceName, XenTime FROM XenProjectXenStoreBase"
	baseDispatchList, err := wmi.QueryNamespaceRaw(baseQuery, &baseList, "root\\wmi")
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		for _, item := range baseDispatchList {
			item.Release()
		}
	}

	log.Debugf("GetBase: Got base list %+v", baseList)

	if len(baseList) != len(baseDispatchList) {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetBase: unexpected length mismatch (baseList=%d, baseDispatchList=%d)",
			len(baseList),
			len(baseDispatchList))
	} else if len(baseList) == 0 {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetBase: Got 0-length list")
	} else if len(baseList) > 1 {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetBase: Unexpected multiple XenProjectXenStoreBase returned")
	}

	return &XenProjectXenStoreBase{
		transport:  t,
		disp:       baseDispatchList[0],
		Properties: baseList[0],
	}, nil
}

func (t *WinPVTransport) GetSession(sessionId int32) (*XenProjectXenStoreSession, error) {
	log.Debug("GetSession: Requesting session list")

	var sessionList []XenProjectXenStoreSessionProps
	sessionQuery := fmt.Sprintf("SELECT Active, ID, InstanceName, SessionID FROM XenProjectXenStoreSession WHERE SessionId=%d", sessionId)
	sessionDispatchList, err := wmi.QueryNamespaceRaw(sessionQuery, &sessionList, "root\\wmi")
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		for _, item := range sessionDispatchList {
			item.Release()
		}
	}

	log.Debugf("GetSession: Got session list: %+v", sessionList)

	if len(sessionList) != len(sessionDispatchList) {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetSession: unexpected length mismatch (sessionList=%d, sessionDispatchList=%d)",
			len(sessionList),
			len(sessionDispatchList))
	} else if len(sessionList) == 0 {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetSession: Got 0-length list")
	} else if len(sessionList) > 1 {
		cleanup()
		return nil, fmt.Errorf("WinPVTransport.GetSession: Unexpected multiple XenProjectXenStoreBase returned")
	}

	return &XenProjectXenStoreSession{
		transport:  t,
		disp:       sessionDispatchList[0],
		Properties: sessionList[0],
	}, nil
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

	session, err := base.AddSession("JoelSession")
	if err != nil {
		return err
	}

	value, err := session.GetValue("vm")
	if err != nil {
		return err
	}

	fmt.Println(value)

	children, err := session.GetChildren(value)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", children)

	if err := session.EndSession(); err != nil {
		return err
	}

	return nil
}
