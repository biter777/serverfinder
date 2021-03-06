package semaphore

import (
	"biterpkg/DEBUG"
	"sync"
	"time"
)

var lockTimeout = time.Minute * 10

// Semaphore - weightless implementation of Semaphore for multithreading, based on "chan struct{}"
type Semaphore struct {
	ch        chan struct{}
	waitPause time.Duration
	mu        *sync.RWMutex
}

// NewSemaphore - create a Semaphore
func NewSemaphore(threads int, waitPause time.Duration) *Semaphore {
	if waitPause < time.Millisecond*100 {
		waitPause = time.Millisecond * 100
	}
	return &Semaphore{
		ch:        make(chan struct{}, threads),
		waitPause: waitPause,
		mu:        &sync.RWMutex{},
	}
}

// Close - close Semaphore
func (s *Semaphore) Close() {
	s.mu.Lock()
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
	s.mu.Unlock()
}

// Lock - lock @n threads of Semaphore
func (s *Semaphore) Lock(n int) {
	var start time.Time
	if DEBUG.ON {
		start = time.Now()
	}
	if n < 1 && s.ch != nil {
		panic("Semaphore::Lock: n < 1")
	}
	for i := 0; i < n && s.ch != nil; i++ {
		s.mu.RLock()
		select {
		case s.ch <- struct{}{}:
		case <-time.After(time.Second * 5):
			i--
		}
		s.mu.RUnlock()
	}
	if DEBUG.ON && time.Now().Sub(start) > lockTimeout {
		panic("Semaphore::Lock: DEBUG ERROR: lockTimeout")
	}
}

// TryLock - try lock @n threads of Semaphore
func (s *Semaphore) TryLock(n int) (success bool, locked int) {
	var start time.Time
	if DEBUG.ON {
		start = time.Now()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ch == nil {
		return false, 0
	}
	if n < 1 && s.ch != nil {
		panic("Semaphore::TryLock: n < 1")
	}
	for i := 0; i < n && s.ch != nil; i++ {
		select {
		case s.ch <- struct{}{}:
			locked++
		default:
			return false, locked
		}
	}
	if DEBUG.ON && time.Now().Sub(start) > lockTimeout {
		panic("Semaphore::TryLock: DEBUG ERROR: lockTimeout")
	}

	return locked >= n, locked
}

// Unlock - unlock @n threads of Semaphore
func (s *Semaphore) Unlock(n int) {
	if n < 1 && s.ch != nil {
		panic("Semaphore::Unlock: n < 1")
	}
	if n > len(s.ch) && s.ch != nil {
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if n > len(s.ch) && s.ch != nil {
				panic("Semaphore::Unlock: n > len(s.ch)")
			}
		}()
	}

	for i := 0; i < n && s.ch != nil; i++ {
		<-s.ch
	}
}

// Len - current lenght of Semaphore chan (numbers of active locks)
func (s *Semaphore) Len() int {
	// s.mu.RLock()
	// defer s.mu.RUnlock()
	return len(s.ch)
}

// Cap - capacity of Semaphore (total max limit of locks)
func (s *Semaphore) Cap() int {
	// s.mu.RLock()
	// defer s.mu.RUnlock()
	return cap(s.ch)
}

// Wait - wait for len(s.ch) will be low than @n
// Example: s.Wait(0) = wait for empty of Semaphore
func (s *Semaphore) Wait(n int) {
	time.Sleep(time.Millisecond * 10)
	if s.ch == nil || len(s.ch) <= n {
		return
	}
	time.Sleep(time.Millisecond + s.waitPause/10)
	for s.ch != nil && len(s.ch) > n {
		time.Sleep(s.waitPause)
	}
}

type isRunninger interface {
	IsRunning() bool
}

// WaitUntilRunning - wait for len(s.ch) will be low than @n
// Example: s.Wait(0) = wait for empty of Semaphore
func (s *Semaphore) WaitUntilRunning(n int, isRunning isRunninger) {
	time.Sleep(time.Millisecond * 10)
	if s.ch == nil || len(s.ch) <= n || !isRunning.IsRunning() {
		return
	}
	time.Sleep(s.waitPause / 10)
	for s.ch != nil && len(s.ch) > n && isRunning.IsRunning() {
		time.Sleep(s.waitPause)
	}
}

// IsFull - true? if semaphore is full (work only if Semaphore.Cap()>0)
func (s *Semaphore) IsFull() bool {
	// s.mu.RLock()
	// defer s.mu.RUnlock()
	l := len(s.ch)
	return l > 0 && l >= cap(s.ch)
}

// IsClosed - IsClosed
func (s *Semaphore) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ch == nil
}

// Chan - возвращает сам канал
func (s *Semaphore) Chan() chan struct{} {
	return s.ch
}

// SetCap - SetCap
// Уменьшение емкости может, по идее, вызвать мертвую блокировку
func (s *Semaphore) SetCap(newCap int) {
	if s == nil || s.ch == nil /*|| cap(s.ch) >= newCap*/ {
		return
	}
	s.mu.RLock()
	// if cap(s.ch) >= newCap {
	// 	s.mu.Unlock()
	// 	return
	// }

	l := len(s.ch)
	if s.ch != nil {
		close(s.ch)
	}
	s.ch = make(chan struct{}, newCap)
	for i := 0; i < l; i++ {
		s.ch <- struct{}{}
	}
	s.mu.Unlock()
}
