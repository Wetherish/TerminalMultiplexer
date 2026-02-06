package session

import (
	"bytes"
	"os"
	"sync"
)

type Session struct {
	ID     int
	Ptmx   *os.File
	Buffer *bytes.Buffer
	Mu     sync.Mutex
}

func (s *Session) Close() {
	if s.Ptmx != nil {
		s.Ptmx.Close()
	}
}
