package xenstore

import (
	"bytes"
	"strings"
)

type Client struct {
	transport Transport
	router    *Router
}

func NewUnixSocketClient(path string) (*Client, error) {
	t, err := NewUnixSocketTransport(path)
	if err != nil {
		return nil, err
	}

	return NewClient(t), nil
}

func NewXenBusClient(path string) (*Client, error) {
	t, err := NewXenBusTransport(path)
	if err != nil {
		return nil, err
	}

	return NewClient(t), nil
}

func NewClient(t Transport) *Client {
	c := &Client{
		transport: t,
		router:    NewRouter(t),
	}

	// Run router in separate goroutine
	go c.router.Loop()

	return c
}

func (c *Client) Close() error {
	c.router.Stop()
	return c.transport.Close()
}

// submitBytes submits a Packet to XenStore and reads a Packet in reply. The response packet
// is checked for errors which are returned from XenStore as strings.
//
// This method blocks until the reply packet is received.
func (c *Client) submitBytes(op xenStoreOperation, payload []byte, txid uint32) (*Packet, error) {
	p, err := NewPacket(op, []byte(payload), 0x0)
	if err != nil {
		return nil, err
	}

	ch, err := c.router.Send(p)
	if err != nil {
		return nil, err
	}

	rsp := <-ch

	if rsp.Header.Op == XsError {
		trimmed := strings.Trim(string(rsp.Payload), "\x00")
		return nil, Error(trimmed)
	}

	return rsp, nil
}

func (c *Client) List(path string) ([]string, error) {
	p, err := c.submitBytes(XsDirectory, []byte(path), 0x0)
	if err != nil {
		return []string{}, err
	}

	// Contents are delimited by NUL bytes
	return strings.Split(p.payloadString(), "\x00"), nil
}

func (c *Client) Read(path string) (string, error) {
	p, err := c.submitBytes(XsRead, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) Remove(path string) (string, error) {
	p, err := c.submitBytes(XsRm, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) Write(path, value string) (string, error) {
	buf := bytes.NewBufferString(path)
	buf.WriteByte(NUL)
	buf.WriteString(value)

	p, err := c.submitBytes(XsWrite, buf.Bytes(), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) GetPermissions(path string) (string, error) {
	p, err := c.submitBytes(XsGetPermissions, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) SetPermissions(path string, perms []string) (string, error) {
	buf := bytes.NewBufferString(path)
	for _, perm := range perms {
		buf.WriteByte(NUL)
		buf.WriteString(perm)
	}

	p, err := c.submitBytes(XsSetPermissions, buf.Bytes(), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) GetDomainPath(path string) (string, error) {
	p, err := c.submitBytes(XsGetDomainPath, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

func (c *Client) Watch(path, token string) (chan *Packet, error) {
	buf := bytes.NewBufferString(path)
	buf.WriteByte(NUL)
	buf.WriteString(token)

	p, err := NewPacket(XsWatch, buf.Bytes(), 0x0)
	if err != nil {
		return nil, err
	}

	return c.router.Send(p)
}

func (c *Client) UnWatch(path, token string) error {
	buf := bytes.NewBufferString(path)
	buf.WriteByte(NUL)
	buf.WriteString(token)

	p, err := c.submitBytes(XsUnWatch, buf.Bytes(), 0x0)
	if err != nil {
		return err
	}

	c.router.removeChannel(p.Header.RqId)

	return nil
}
