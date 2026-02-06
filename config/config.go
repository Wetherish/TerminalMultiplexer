package config

import (
	"fmt"
)

type Shortcut struct {
	Name   string
	Action func()
}

var Config = map[byte]Shortcut{
	15: { // Ctrl + o
		Name: "Open Menu",
		Action: func() {
			fmt.Print("\r\n\033[32m[TaskFlow] Menu Opened!\033[0m\r\n")
		},
	},
	18: { // Ctrl + r
		Name: "Run Project",
		Action: func() {
			fmt.Print("\r\n\033[34m[TaskFlow] Running Project...\033[0m\r\n")
		},
	},
}
