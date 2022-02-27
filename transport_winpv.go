//go:build windows
// +build windows

package xenstore

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/yusufpapurcu/wmi"
)

type WinPVTransport struct {
	base       *XenProjectXenStoreBase
	session    *XenProjectXenStoreSession
	pktChannel chan *Packet
}

func (t *WinPVTransport) Close() error {
	wmi.DefaultClient.SWbemServicesClient.Close()

	if t.base != nil {
		t.base.disp.Release()
	}

	if t.session != nil {
		if err := t.session.EndSession(); err != nil {
			return err
		}
	}

	return nil
}

func (t *WinPVTransport) Send(pkt *Packet) error {
	packet := &Packet{
		Header: &PacketHeader{
			Op:   pkt.Header.Op,
			RqId: pkt.Header.RqId,
			TxId: pkt.Header.TxId,
		},
	}

	if pkt.Header.Op == XsRead {
		val, err := t.session.GetValue(string(pkt.Payload))
		if err != nil {
			return err
		}
		packet.Header.Length = uint32(len(val))
		packet.Payload = []byte(val)
	}

	t.pktChannel <- packet

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
	if t.session != nil {
		return t.session, nil
	}

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

	t.session = &XenProjectXenStoreSession{
		transport:  t,
		disp:       sessionDispatchList[0],
		Properties: sessionList[0],
	}

	return t.session, nil
}

func NewWinPVTransport() (*WinPVTransport, error) {
	if err := initWmi(); err != nil {
		return nil, err
	}

	transport := &WinPVTransport{}

	c := make(chan *Packet)
	transport.pktChannel = c

	base, err := transport.GetBase()
	if err != nil {
		return nil, err
	}

	session, err := base.AddSession("JoelSession")
	if err != nil {
		return nil, err
	}
	transport.session = session

	value, err := session.GetValue("vm")
	if err != nil {
		return nil, err
	}

	fmt.Println(value)

	children, err := session.GetChildren(value)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", children)

	return transport, nil
}
