package xenstore

import (
	"bytes"
	"strconv"
	"strings"
)

// Client is a wrapper which allows easier communication with XenStore by providing
// methods which allow performing normal XenStore functions with minimal effort.
type Client struct {
	transport Transport
	router    *Router
	stopError error
}

// NewUnixSocketClient creates a new Client which will be connected to an underlying
// UnixSocket.
func NewUnixSocketClient(path string) (*Client, error) {
	t, err := NewUnixSocketTransport(path)
	if err != nil {
		return nil, err
	}

	return NewClient(t), nil
}

// NewXenBusClient creates a new Client which will be connected to an underlying
// XenBus device.
func NewXenBusClient(path string) (*Client, error) {
	t, err := NewXenBusTransport(path)
	if err != nil {
		return nil, err
	}

	return NewClient(t), nil
}

// NewClient creates a new connected Client and starts the internal Router so
// that packets can be sent and received correctly by the Client.
func NewClient(t Transport) *Client {
	c := &Client{
		transport: t,
		router:    NewRouter(t),
	}

	// Run router in separate goroutine
	go func() {
		c.stopError = c.router.Start()
	}()

	return c
}

// Close stops the underlying Router and closes the Transport.
func (c *Client) Close() error {
	c.router.Stop()
	return c.transport.Close()
}

func (c *Client) Error() error {
	return c.stopError
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

// List lists the descendants of path.
func (c *Client) List(path string) ([]string, error) {
	p, err := c.submitBytes(XsDirectory, []byte(path), 0x0)
	if err != nil {
		return []string{}, err
	}

	// Contents are delimited by NUL bytes
	return strings.Split(p.payloadString(), "\x00"), nil
}

// Read reads the contents of path from XenStore.
func (c *Client) Read(path string) (string, error) {
	p, err := c.submitBytes(XsRead, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

// Remove removes a path from XenStore recursively
func (c *Client) Remove(path string) (string, error) {
	p, err := c.submitBytes(XsRm, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

// Write value to XenStore at path.
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

// GetPermissions returns the currently stored permissions for a XenStore path.
func (c *Client) GetPermissions(path string) (string, error) {
	p, err := c.submitBytes(XsGetPermissions, []byte(path), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

// SetPermissions sets the permissions for a path in XenStore.
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

// GetDomainPath
func (c *Client) GetDomainPath(domid int) (string, error) {
	s := strconv.Itoa(domid)

	p, err := c.submitBytes(XsGetDomainPath, []byte(s), 0x0)
	if err != nil {
		return "", err
	}

	return p.payloadString(), nil
}

// Watch places a watch on a particular XenStore path
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

// UnWatch removes a previously-set watch on a XenStore path.
func (c *Client) UnWatch(path, token string) error {
	buf := bytes.NewBufferString(path)
	buf.WriteByte(NUL)
	buf.WriteString(token)

	p, err := c.submitBytes(XsUnWatch, buf.Bytes(), 0x0)
	if err != nil {
		return err
	}

	// Ensure the returned packet was not an error
	if err := p.Check(); err != nil {
		return err
	}

	c.router.removeWatchChannel(token)

	return nil
}
