package serverfinder

import (
	"errors"
	"sync"
	"time"
)

type responseChan struct {
	sync.Mutex
	ch     chan *response
	closed bool
}

type response struct {
	err          error
	port         int // основной порт сервера
	portOptional int // дополнительный порт сервера, если поддерживается, опционально
}

// ------------------------------------------------------------------

func newResponseChan(cap int) *responseChan {
	return &responseChan{ch: make(chan *response, cap)}
}

// ------------------------------------------------------------------

func (c *responseChan) close() {
	if c == nil {
		return
	}
	c.Lock()
	if c.ch != nil && !c.closed {
		c.closed = true
		close(c.ch)
	}
	c.Unlock()
}

// ------------------------------------------------------------------

func (c *responseChan) wait() {
	for !c.closed && c.isFull() {
		time.Sleep(time.Millisecond)
	}
}

// ------------------------------------------------------------------

func (c *responseChan) send(resp *response) error {
	c.wait()
	c.Lock()
	defer c.Unlock()
	if c.ch == nil || c.closed {
		return errors.New("chan is closed")
	}
	if len(c.ch) >= cap(c.ch) {
		return c.send(resp)
	}
	c.ch <- resp
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
