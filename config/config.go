package config

type Shortcut struct {
	Name   string
	Action func(t any)
}

func NewShortcut(name string, action func(t any)) Shortcut {
	return Shortcut{Name: name, Action: action}
}

type Config struct {
	shortcuts map[string]Shortcut
}

func NewConfig() *Config {
	return &Config{
		shortcuts: make(map[string]Shortcut),
	}
}

func (c *Config) RegisterShortcut(key string, shortcut Shortcut) {
	c.shortcuts[key] = shortcut
}

func (c *Config) FindAndInvoke(key string, t any) bool {
	value, ok := c.shortcuts[key]
	if ok {
		value.Action(t)
		return true
	}
	return false
}
