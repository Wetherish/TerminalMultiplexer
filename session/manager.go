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
	Config    *config.Config
	nextID    int
	winSize   *pty.Winsize
}

func MakeAction[T any](f func(T)) func(any) {
	return func(a any) {
		if val, ok := a.(T); ok {
			f(val)
		}
	}
}

func NewManager() *Manager {
	mgr := &Manager{
		Sessions:  make([]*Session, 0),
		ActiveIdx: 0,
		Config:    config.LoadConfig(),
		nextID:    1,
	}

	switchTab := MakeAction(func(m *Manager) {
		m.SwitchTab()
	})

	newTab := MakeAction(func(m *Manager) {
		m.AddSession()
		m.Draw()
	})

	// loadNewConfig := MakeAction(func(m *Manager) {
	// 	m.Config = config.LoadConfig()
	// 	m.Draw()
	// })

	mgr.Config.RegisterShortcut(config.CtrlN, config.NewShortcut("New Tab", newTab))
	// mgr.Config.RegisterShortcut(config.CtrlB, config.NewShortcut("Load New Config", loadNewConfig))
	mgr.Config.RegisterShortcut(config.ShiftTab, config.NewShortcut("Switch Tab", switchTab))

	return mgr
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
		Buffer: NewRingBuffer(1024 * 1024),
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

		data := buf[:n]

		s.Mu.Lock()
		s.Buffer.Write(data)
		s.Mu.Unlock()

		hasClear := bytes.Contains(data, clearSeq)

		m.Mu.Lock()
		isActive := len(m.Sessions) > 0 && m.Sessions[m.ActiveIdx].ID == s.ID

		if isActive {
			os.Stdout.Write(data)
			if hasClear {
				m.Mu.Unlock()
				m.Draw()
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
	m.Draw()
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
			m.Draw()
			m.Mu.Lock()
			return
		}
	}
}

func (m *Manager) Draw() {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return
	}

	m.Mu.Lock()
	sessions := m.Sessions
	activeIdx := m.ActiveIdx
	conf := m.Config.UI
	m.Mu.Unlock()
	var bar string
	if conf.EnableSideBar {
		bar = RenderSideBar(conf, sessions, activeIdx, width)
	} else {
		bar = RenderFooter(conf, sessions, activeIdx, width)
	}

	fmt.Print("\0337")
	region := fmt.Sprintf("\033[1;%dr", height-1)
	fmt.Print(region)

	m.Mu.Lock()
	m.winSize = &pty.Winsize{Rows: uint16(height - 1), Cols: uint16(width)}
	for _, s := range m.Sessions {
		pty.Setsize(s.Ptmx, m.winSize)
	}
	m.Mu.Unlock()

	if conf.EnableSideBar {
		fmt.Printf("\033[1;1f\033[2K%s", bar)
	} else {
		fmt.Printf("\033[%d;1f\033[2K%s", height, bar)
	}
	fmt.Print("\0338")
}

func (m *Manager) Run() {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			m.Draw()
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

		if m.Config.FindAndInvoke(string(input), m) {
			continue
		}

		if len(input) == 1 && input[0] == 17 {
			return
		}

		m.Mu.Lock()
		if len(m.Sessions) > 0 {
			m.Sessions[m.ActiveIdx].Ptmx.Write(input)
		}
		m.Mu.Unlock()
	}
}
