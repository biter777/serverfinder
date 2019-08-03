package serverfinder

import (
	"errors"
	"sync"
)

type responseChan struct {
	sync.Mutex
	ch chan *response
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
	if c.ch != nil {
		close(c.ch)
		c.ch = nil
	}
	c.Unlock()
}

// ------------------------------------------------------------------

func (c *responseChan) send(resp *response) error {
	c.Lock()
	if c.ch == nil {
		return errors.New("chan is nil")
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
