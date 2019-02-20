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

	*Grid
	dirty       chan struct{} // update cached buffer
	die         chan struct{} // client disconnect

	debugPipe   net.Conn
	debug       bool
}

func NewClient(p1 net.Conn, user *Grid) (*Client) {
	c := &Client{
		Grid: user,
		Conn: p1,
		dirty: make(chan struct{}, 1),
		die: make(chan struct{}),
	}
	go c.worker()

	if *verbosity > 4 {
		rx, tx := net.Pipe()
		c.debugPipe = tx
		c.debug = true

		go func (p1 net.Conn) {
			buffer := make([]byte, (1<<16)+headerSize)

			for {
				f, err := readFrame(p1, buffer)
				if err != nil {
					return
				}
				Vln(5, "[dump]", f)
			}
		}(rx)
	}

	return c
}

func (c *Client) Read(buff []byte) (int, error) {
	c.rmx.Lock()
	defer c.rmx.Unlock()
	return c.Conn.Read(buff)
}

func (c *Client) Write(buff []byte) (int, error) {
	c.wmx.Lock()
	defer c.wmx.Unlock()

	if c.debug {
		c.debugPipe.Write(buff)
	}
	return c.Conn.Write(buff)
}

func (c *Client) ReadFrame(buffer []byte) (f Frame, err error) {
	c.rmx.Lock()
	defer c.rmx.Unlock()
	return readFrame(c.Conn, buffer)
}


func (c *Client) WriteRawFrame(dataHex string) (n int, err error) {
	//c.wmx.Lock()
	//defer c.wmx.Unlock()
	return writeRawFrame(c, dataHex)
}

func (c *Client) WriteFrame(data []byte) (n int, err error) {
	//c.wmx.Lock()
	//defer c.wmx.Unlock()
	return writeFrame(c, data)
}

func (c *Client) Flush() {
	select {
	case c.dirty <- struct{}{}:
	default:
	}
}

func (c *Client) worker() {
	for {
		select {
		case <-c.dirty:
			c.Grid.BuildCachedAll()
			c.Grid.BuildCached()

			// resend page data
			Vln(4, "[resend]", len(c.Grid.buf))
			c.WriteAllPage()

		case <-c.die:
			return
		}
	}
}

func (c *Client) Close() (error) {
	select {
	case <-c.die:
		return nil
	default:
		close(c.die)
		return c.Conn.Close()
	}
}

func (c *Client) WritePage(page int) {
	buf := c.GetPage(page)
	Vf(5, "[page]%d, %d, % 02X\n", page, len(buf), buf)
	if buf != nil {
		size := len(buf)
		head := Raw2Byte("26 08 85 35 00 00 06")
		head[6] = uint8(size)
		//Vf(4, "[page]%d, % 02X\n", page, head)
		for i := 0; i < size; i++ {
			head = append(head, buf[i]...)
		}
		c.WriteFrame(head)
	}
}

func (c *Client) WriteAllPage() {
	pages := c.GetAllPage()
	Vf(5, "[pages]%v\n", pages)
	for idx, _ := range pages {
		c.WritePage(int(idx))
	}
}
