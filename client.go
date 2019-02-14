package main

import (
	//"encoding/binary"
	//"fmt"
	//"io"
	"net"
	"sync"
	//"time"
)


type Client struct {
	net.Conn
	rmx         sync.Mutex
	wmx         sync.Mutex

	User        *User
}

func NewClient(p1 net.Conn) (*Client) {
	return &Client{
		Conn: p1,
	}
}

func (c *Client) Read(buff []byte) (int, error) {
	c.rmx.Lock()
	defer c.rmx.Unlock()

	return c.Conn.Read(buff)
}

func (c *Client) Write(buff []byte) (int, error) {
	c.wmx.Lock()
	defer c.wmx.Unlock()

	return c.Conn.Write(buff)
}

func (c *Client) ReadFrame(buffer []byte) (f Frame, err error) {
	c.rmx.Lock()
	defer c.rmx.Unlock()

	return readFrame(c.Conn, buffer)
}


func (c *Client) WriteRawFrame(dataHex string) (n int, err error) {
	c.wmx.Lock()
	defer c.wmx.Unlock()

	return writeRawFrame(c.Conn, dataHex)
}

func (c *Client) WriteFrame(data []byte) (n int, err error) {
	c.wmx.Lock()
	defer c.wmx.Unlock()

	return writeFrame(c.Conn, data)
}
