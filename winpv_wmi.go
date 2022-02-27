//go:build windows
// +build windows

package xenstore

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	log "github.com/sirupsen/logrus"
	"github.com/yusufpapurcu/wmi"
)

func initWmi() error {
	swb, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		return err
	}
	wmi.DefaultClient.SWbemServicesClient = swb
	return nil
}

type XenProjectXenStoreBase struct {
	transport  *WinPVTransport
	disp       *ole.IDispatch
	Properties XenProjectXenStoreBaseProps
}

type XenProjectXenStoreBaseProps struct {
	Active       bool
	InstanceName string
	XenTime      uint64
}

type XenProjectXenStoreSession struct {
	transport  *WinPVTransport
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

func (b *XenProjectXenStoreBase) AddSession(name string) (*XenProjectXenStoreSession, error) {
	log.Debug("AddSession: Adding session")

	sessionResultRaw := new(ole.VARIANT)
	ole.VariantInit(sessionResultRaw)
	defer sessionResultRaw.Clear()

	methodName := "AddSession"
	resultRaw, err := oleutil.CallMethod(b.disp, methodName, name, sessionResultRaw)
	if err != nil {
		return nil, fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	log.Debugf("AddSession: Added session: %+v", sessionResultRaw)

	sessionId := sessionResultRaw.Value().(int32)

	return b.transport.GetSession(sessionId)
}

func (s *XenProjectXenStoreSession) GetValue(path string) (string, error) {
	log.Debugf("GetValue: Getting value for path: %s", path)

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

func (s *XenProjectXenStoreSession) GetChildren(path string) ([]string, error) {
	log.Debugf("GetChildren: Getting children for path: %s", path)

	var stringNodes []string

	valueResultRaw := new(ole.VARIANT)
	ole.VariantInit(valueResultRaw)
	defer valueResultRaw.Clear()

	methodName := "GetChildren"
	resultRaw, err := oleutil.CallMethod(s.disp, methodName, path, valueResultRaw)
	if err != nil {
		return stringNodes, fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	children := valueResultRaw.ToIDispatch()
	defer children.Release()

	nodeResult, err := oleutil.GetProperty(children, "ChildNodes")
	if err != nil {
		return stringNodes, err
	}
	defer nodeResult.Clear()

	for _, node := range nodeResult.ToArray().ToValueArray() {
		stringNodes = append(stringNodes, node.(string))
	}

	return stringNodes, nil
}

// EndSession ends the session and cleans up the reference to it. The session must not be used after calling this method.
func (s *XenProjectXenStoreSession) EndSession() error {
	log.Debugf("EndSession: Ending session: %d", s.Properties.SessionID)

	defer s.disp.Release()

	methodName := "EndSession"
	resultRaw, err := oleutil.CallMethod(s.disp, methodName)
	if err != nil {
		return fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	return nil
}
