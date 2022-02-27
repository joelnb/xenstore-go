//go:build windows
// +build windows

package xenstore

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/joelnb/wmi"
	log "github.com/sirupsen/logrus"
)

const SessionName = "xenstore-go"

type WinPVTransport struct {
	base       *XenProjectXenStoreBase
	pktChannel chan *Packet

	Session *XenProjectXenStoreSession
}

func (t *WinPVTransport) Close() error {
	wmi.DefaultClient.SWbemServicesClient.Close()

	if t.base != nil {
		t.base.disp.Release()
	}

	return nil
}

func (t *WinPVTransport) Send(pkt *Packet) error {
	var err error

	session := t.Session
	if t.Session == nil {
		session, err = t.base.AddSession(SessionName)
		if err != nil {
			return err
		}
	}

	rpkt := &Packet{
		Header: &PacketHeader{
			Op:   pkt.Header.Op,
			RqId: pkt.Header.RqId,
			TxId: pkt.Header.TxId,
		},
	}

	switch pkt.Header.Op {
	case XsDirectory:
		val, err := session.GetChildren(string(pkt.Payload))
		if err != nil {
			return err
		}

		rpkt.Payload = []byte(strings.Join(val, "\000"))
	case XsRead:
		val, err := session.GetValue(string(pkt.Payload))
		if err != nil {
			return err
		}

		rpkt.Payload = []byte(val)
	case XsRm:
		if err := session.RemoveValue(string(pkt.Payload)); err != nil {
			return err
		}
	case XsWrite:
		args := pkt.Strings()
		if err := session.SetValue(args[0], args[1]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("WinPVTransport: Unsupported packet: %+v", pkt)
	}

	rpkt.Header.Length = uint32(len(rpkt.Payload))

	t.pktChannel <- rpkt

	if t.Session == nil {
		if err := session.EndSession(); err != nil {
			return err
		}
	}

	return nil
}

func (t *WinPVTransport) Receive() (*Packet, error) {
	rsp := <-t.pktChannel
	return rsp, nil
}

func (t *WinPVTransport) GetBase() (*XenProjectXenStoreBase, error) {
	if t.base != nil {
		return t.base, nil
	}

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

	t.base = &XenProjectXenStoreBase{
		transport:  t,
		disp:       baseDispatchList[0],
		Properties: baseList[0],
	}

	return t.base, nil
}

func (t *WinPVTransport) GetSession(sessionId int32) (*XenProjectXenStoreSession, error) {
	sessionList, sessionDispatchList, err := t.getSessions(sessionId)
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		for _, item := range sessionDispatchList {
			item.Release()
		}
	}

	if len(sessionList) == 0 {
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

func (t *WinPVTransport) GetSessions() ([]*XenProjectXenStoreSession, error) {
	var sessions []*XenProjectXenStoreSession

	sessionList, sessionDispatchList, err := t.getSessions()
	if err != nil {
		return nil, err
	}

	for i := range sessionList {
		sessions = append(sessions, &XenProjectXenStoreSession{
			transport:  t,
			disp:       sessionDispatchList[i],
			Properties: sessionList[i],
		})
	}

	return sessions, nil
}

func (t *WinPVTransport) getSessions(sessionIds ...int32) ([]XenProjectXenStoreSessionProps, []*ole.IDispatch, error) {
	log.Debug("WinPVTransport.getSessions: Requesting session list")

	sessionQuery := "SELECT Active, ID, InstanceName, SessionID FROM XenProjectXenStoreSession"
	if len(sessionIds) > 0 {
		sessionQuery = fmt.Sprintf("%s WHERE SessionId=%d", sessionQuery, sessionIds[0])
	}

	var sessionList []XenProjectXenStoreSessionProps
	sessionDispatchList, err := wmi.QueryNamespaceRaw(sessionQuery, &sessionList, "root\\wmi")
	if err != nil {
		return []XenProjectXenStoreSessionProps{}, []*ole.IDispatch{}, err
	}

	log.Debugf("WinPVTransport.getSessions: Got session list: %+v", sessionList)

	if len(sessionList) != len(sessionDispatchList) {
		for _, item := range sessionDispatchList {
			item.Release()
		}

		return []XenProjectXenStoreSessionProps{}, []*ole.IDispatch{}, fmt.Errorf("WinPVTransport.getSessions: unexpected length mismatch (sessionList=%d, sessionDispatchList=%d)",
			len(sessionList),
			len(sessionDispatchList))
	}

	return sessionList, sessionDispatchList, nil
}

func NewWinPVTransport() (*WinPVTransport, error) {
	if err := initWmi(); err != nil {
		return nil, err
	}

	transport := &WinPVTransport{}

	c := make(chan *Packet)
	transport.pktChannel = c

	_, err := transport.GetBase()
	if err != nil {
		return nil, err
	}

	return transport, nil
}
