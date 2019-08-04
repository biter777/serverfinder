package serverfinder

import (
	"errors"
	"sync"
	"time"
)

type responseChan struct {
	sync.Mutex
	ch       chan *response
	chClosed bool
}

type response struct {
	err  error
	port int
}

// ------------------------------------------------------------------

func newResponseChan(cap int) *responseChan {
	return &responseChan{ch: make(chan *response, cap)}
}

// ------------------------------------------------------------------

func (c *responseChan) close() {
	c.Lock()
	if c.ch != nil && !c.chClosed {
		c.chClosed = true
		close(c.ch)
	}
	c.Unlock()
}

// ------------------------------------------------------------------

func (c *responseChan) wait() {
	for !c.chClosed && c.isFull() {
		time.Sleep(time.Millisecond)
	}
}

// ------------------------------------------------------------------

func (c *responseChan) send(resp *response) error {
	c.wait()
	c.Lock()
	if c.ch == nil || c.chClosed {
		c.Unlock()
		return errors.New("chan is closed")
	}
	if len(c.ch) >= cap(c.ch) {
		c.Unlock()
		return c.send(resp)
	}
	c.ch <- resp
	c.Unlock()
	return nil
}

// ------------------------------------------------------------------

func (c *responseChan) rcv() *response {
	return <-c.ch
}

// ------------------------------------------------------------------

func (c *responseChan) isFull() bool {
	c.Lock()
	full := len(c.ch) >= cap(c.ch)
	c.Unlock()
	return full
}

// ------------------------------------------------------------------
