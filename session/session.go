package session

import (
	"os"
	"sync"
)

type Session struct {
	ID     int
	Ptmx   *os.File
	Buffer *RingBuffer
	Mu     sync.Mutex
}

func (s *Session) Close() {
	if s.Ptmx != nil {
		s.Ptmx.Close()
	}
}
