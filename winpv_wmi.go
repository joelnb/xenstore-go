//go:build windows
// +build windows

/*
Some examples of how to use the PowerShell API

```
# Remove all sessions & confirm they are gone
$sessions = Get-WmiObject -Namespace root\wmi -Class XenProjectXenStoreSession
$sessions | ForEach-Object { $_.EndSession() }
Get-WmiObject -Namespace root\wmi -Class XenProjectXenStoreSession

# Get a value & list a directory
$base = Get-WmiObject -Namespace root\wmi -Class XenProjectXenStoreBase -ErrorAction SilentlyContinue
$sid = $base.AddSession("ExampleSession")
$session = Get-WmiObject -Namespace root\wmi -Query "select * from XenProjectXenStoreSession where SessionId=$($sid.SessionId)"
$my_path = $session.GetValue("vm")['value']
$my_nodes = $session.GetChildren($my_path)
$my_nodes.children.ChildNodes
$session.EndSession()
```
*/

package xenstore

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/joelnb/wmi"
	log "github.com/sirupsen/logrus"
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
	transport *WinPVTransport
	disp      *ole.IDispatch

	Properties XenProjectXenStoreBaseProps
}

type XenProjectXenStoreBaseProps struct {
	Active       bool
	InstanceName string
	XenTime      uint64
}

type XenProjectXenStoreSession struct {
	transport *WinPVTransport
	disp      *ole.IDispatch

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

func resultVariant() *ole.VARIANT {
	variant := new(ole.VARIANT)
	ole.VariantInit(variant)
	return variant
}

func (s *XenProjectXenStoreSession) callMethod(name string, params ...interface{}) (*ole.VARIANT, error) {
	resultRaw, err := oleutil.CallMethod(s.disp, name, params...)
	if err != nil {
		return nil, fmt.Errorf("CallMethod XenProjectXenStoreSession.%s: %v", name, err)
	}
	return resultRaw, nil
}

func (b *XenProjectXenStoreBase) AddSession(name string) (*XenProjectXenStoreSession, error) {
	log.Debug("AddSession: Adding session")

	sessionResultRaw := resultVariant()
	defer sessionResultRaw.Clear()

	methodName := "AddSession"
	resultRaw, err := oleutil.CallMethod(b.disp, methodName, name, sessionResultRaw)
	if err != nil {
		return nil, fmt.Errorf("CallMethod XenProjectXenStoreBase.%s: %v", methodName, err)
	}
	defer resultRaw.Clear()

	log.Debugf("AddSession: Added session: %+v", sessionResultRaw)

	return b.transport.GetSession(sessionResultRaw.Value().(int32))
}

func (s *XenProjectXenStoreSession) GetValue(path string) (string, error) {
	log.Debugf("GetValue: Getting value for path: %s", path)

	valueResultRaw := resultVariant()
	defer valueResultRaw.Clear()

	resultRaw, err := s.callMethod("GetValue", path, valueResultRaw)
	if err != nil {
		return "", err
	}
	defer resultRaw.Clear()

	return valueResultRaw.ToString(), nil
}

func (s *XenProjectXenStoreSession) SetValue(path, value string) error {
	log.Debugf("SetValue: Setting value '%s' for path: %s", value, path)

	resultRaw, err := s.callMethod("SetValue", path, value)
	if err != nil {
		return err
	}
	defer resultRaw.Clear()

	return nil
}

func (s *XenProjectXenStoreSession) Log(msg string) error {
	resultRaw, err := s.callMethod("Log", msg)
	if err != nil {
		return err
	}
	defer resultRaw.Clear()

	return nil
}

// GetChildren lists the children of the requested path.
func (s *XenProjectXenStoreSession) GetChildren(path string) ([]string, error) {
	log.Debugf("GetChildren: Getting children for path: %s", path)

	var stringNodes []string

	valueResultRaw := resultVariant()
	defer valueResultRaw.Clear()

	resultRaw, err := s.callMethod("GetChildren", path, valueResultRaw)
	if err != nil {
		return nil, err
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

	resultRaw, err := s.callMethod("EndSession")
	defer resultRaw.Clear()

	return err
}

// RemoveValue removes a Xenstore key.
func (s *XenProjectXenStoreSession) RemoveValue(path string) error {
	log.Debugf("RemoveValue: Removing value: %s", path)

	defer s.disp.Release()

	resultRaw, err := s.callMethod("RemoveValue", path)
	defer resultRaw.Clear()

	return err
}

// SetWatch creates a Xenstore watch.
// TODO: Allow subscribing to the WMI events related to the watch.
func (s *XenProjectXenStoreSession) SetWatch(path string) error {
	log.Debugf("SetWatch: Setting watch on: %s", path)

	defer s.disp.Release()

	resultRaw, err := s.callMethod("SetWatch", path)
	defer resultRaw.Clear()

	return err
}

// RemoveWatch removes a previously set Xenstore watch.
func (s *XenProjectXenStoreSession) RemoveWatch(path string) error {
	log.Debugf("RemoveWatch: Removing watch on: %s", path)

	defer s.disp.Release()

	resultRaw, err := s.callMethod("RemoveWatch", path)
	defer resultRaw.Clear()

	return err
}

// StartTransaction starts a new transaction to atomically group some actions.
func (s *XenProjectXenStoreSession) StartTransaction() error {
	defer s.disp.Release()

	resultRaw, err := s.callMethod("StartTransaction")
	defer resultRaw.Clear()

	return err
}

// CommitTransaction ends the transaction, causing all of the actions within it to be processed.
func (s *XenProjectXenStoreSession) CommitTransaction() error {
	defer s.disp.Release()

	resultRaw, err := s.callMethod("CommitTransaction")
	defer resultRaw.Clear()

	return err
}

// AbortTransaction aborts a transaction, discarding any actions that were done since it was started.
func (s *XenProjectXenStoreSession) AbortTransaction() error {
	defer s.disp.Release()

	resultRaw, err := s.callMethod("AbortTransaction")
	defer resultRaw.Clear()

	return err
}
