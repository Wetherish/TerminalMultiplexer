package main

import (
	"damnTerminal/session"
	"fmt"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
			os.Exit(1)
		}
	}()

	fmt.Println("Starting TaskFlow...")
	mgr := session.NewManager()

	mgr.AddSession()

	mgr.Run()
}

// package main

// import (
// 	"fmt"
// 	"os"

// 	"golang.org/x/term"
// )

// func main() {
// 	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
// 	defer term.Restore(int(os.Stdin.Fd()), oldState)

// 	fmt.Print("Press Ctrl+Tab (or any key) to see its code. Press Ctrl+C to exit.\r\n")

// 	buf := make([]byte, 1024)
// 	for {
// 		n, _ := os.Stdin.Read(buf)
// 		if n > 0 {
// 			fmt.Printf("Bytes: %v  String: %q\r\n", buf[:n], buf[:n])
// 			if buf[0] == 3 { // Ctrl+C
// 				return
// 			}
// 		}
// 	}
// }
