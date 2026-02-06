package session

import (
	"bytes"
	"damnTerminal/config"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type Manager struct {
	Sessions  []*Session
	ActiveIdx int
	Mu        sync.Mutex
	nextID    int
	winSize   *pty.Winsize
}

func NewManager() *Manager {
	return &Manager{
		Sessions:  make([]*Session, 0),
		ActiveIdx: 0,
		nextID:    1,
	}
}

func (m *Manager) AddSession() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	cmd := exec.Command("bash")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("Failed: %s\r\n", err)
		return
	}

	if m.winSize != nil {
		pty.Setsize(ptmx, m.winSize)
	}

	s := &Session{
		ID:     m.nextID,
		Ptmx:   ptmx,
		Buffer: new(bytes.Buffer),
	}
	m.nextID++

	m.Sessions = append(m.Sessions, s)
	m.ActiveIdx = len(m.Sessions) - 1

	go m.streamSessionOutput(s)
}

func (m *Manager) streamSessionOutput(s *Session) {
	buf := make([]byte, 4096)
	clearSeq := []byte("\x1b[2J")

	for {
		n, err := s.Ptmx.Read(buf)
		if err != nil {
			m.RemoveSession(s.ID)
			return
		}

		s.Mu.Lock()
		s.Buffer.Write(buf[:n])
		s.Mu.Unlock()

		hasClear := bytes.Contains(buf[:n], clearSeq)

		m.Mu.Lock()
		isActive := len(m.Sessions) > 0 && m.Sessions[m.ActiveIdx].ID == s.ID

		if isActive {
			os.Stdout.Write(buf[:n])
			if hasClear {
				m.Mu.Unlock()
				m.DrawFooter()
				// Lock again to satisfy the defer/logic flow if we had more code
				// (Though here we just loop, so we don't strictly need to relock
				// unless we access m fields again in this block)
			} else {
				m.Mu.Unlock()
			}
		} else {
			m.Mu.Unlock()
		}
	}
}

func (m *Manager) SwitchTab() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if len(m.Sessions) == 0 {
		return
	}

	m.ActiveIdx++
	if m.ActiveIdx >= len(m.Sessions) {
		m.ActiveIdx = 0
	}

	activeSession := m.Sessions[m.ActiveIdx]

	fmt.Print("\033[2J\033[H")

	activeSession.Mu.Lock()
	os.Stdout.Write(activeSession.Buffer.Bytes())
	activeSession.Mu.Unlock()

	if m.winSize != nil {
		pty.Setsize(activeSession.Ptmx, m.winSize)
	}

	m.Mu.Unlock()
	m.DrawFooter()
	m.Mu.Lock()
}

func (m *Manager) RemoveSession(id int) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	for i, s := range m.Sessions {
		if s.ID == id {
			s.Close()
			m.Sessions = append(m.Sessions[:i], m.Sessions[i+1:]...)
			if len(m.Sessions) == 0 {
				os.Exit(0)
			}
			if m.ActiveIdx >= len(m.Sessions) {
				m.ActiveIdx = len(m.Sessions) - 1
			}

			fmt.Print("\033[2J\033[H")

			sNew := m.Sessions[m.ActiveIdx]
			sNew.Mu.Lock()
			os.Stdout.Write(sNew.Buffer.Bytes())
			sNew.Mu.Unlock()

			m.Mu.Unlock()
			m.DrawFooter()
			m.Mu.Lock()
			return
		}
	}
}

func (m *Manager) DrawFooter() {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return
	}

	fmt.Print("\0337")
	fmt.Printf("\033[1;%dr", height-1)

	m.winSize = &pty.Winsize{Rows: uint16(height - 1), Cols: uint16(width)}
	m.Mu.Lock()
	for _, s := range m.Sessions {
		pty.Setsize(s.Ptmx, m.winSize)
	}

	displayIdx := m.ActiveIdx + 1
	totalSessions := len(m.Sessions)
	m.Mu.Unlock()

	text := fmt.Sprintf("\033[30;42m TaskFlow | Tab: %d/%d | Ctrl+N: New | Ctrl+B: Switch | Ctrl+Q: Quit \033[0m",
		displayIdx, totalSessions)

	fmt.Printf("\033[%d;1f\033[2K%s", height, text)

	fmt.Print("\0338")
}

func (m *Manager) Run() {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			m.DrawFooter()
		}
	}()
	ch <- syscall.SIGWINCH

	buf := make([]byte, 1024)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}

		input := buf[:n]
		ctrlTabSeq := []byte{27, 91, 90}

		if bytes.Equal(input, ctrlTabSeq) {
			m.SwitchTab()
			continue
		}

		if len(input) == 1 {
			key := input[0]

			if key == 14 {
				m.AddSession()
				m.DrawFooter()
				continue
			}
			if key == 17 {
				return
			}
			if shortcut, ok := config.Config[key]; ok {
				shortcut.Action()
				continue
			}
		}

		m.Mu.Lock()
		if len(m.Sessions) > 0 {
			m.Sessions[m.ActiveIdx].Ptmx.Write(input)
		}
		m.Mu.Unlock()
	}
}
