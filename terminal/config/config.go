package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type UIConfig struct {
	BarBackgroundColor       string `toml:"bar_background_color"`
	BarForegroundColor       string `toml:"bar_foreground_color"`
	TabBackgroundColor       string `toml:"tab_background_color"`
	TabForegroundColor       string `toml:"tab_foreground_color"`
	ActiveTabBackgroundColor string `toml:"active_tab_background_color"`
	ActiveTabForegroundColor string `toml:"active_tab_foreground_color"`
	EnableSideBar            bool   `toml:"enable_side_bar"`
}

type ApplicationConfig struct {
}

type KeymapConfig struct {
}

type Config struct {
	Shortcuts map[string]Shortcut
	UI        UIConfig          `toml:"ui"`
	App       ApplicationConfig `toml:"app"`
	Keymap    KeymapConfig      `toml:"keymap"`
}

func LoadConfig() *Config {
	file, err := os.ReadFile("/Users/tymczasowe/Projekt/config.toml")
	if err != nil {
		panic(err)
	}

	var config Config
	if err := toml.Unmarshal(file, &config); err != nil {
		panic(err)
	}
	config.Shortcuts = make(map[string]Shortcut)
	return &config
}

func NewConfig() *Config {
	return &Config{
		Shortcuts: make(map[string]Shortcut),
		UI:        UIConfig{},
		App:       ApplicationConfig{},
		Keymap:    KeymapConfig{},
	}
}
