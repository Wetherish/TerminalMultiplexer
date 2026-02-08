package config

type Shortcut struct {
	Name   string
	Action func(t any)
}

func NewShortcut(name string, action func(t any)) Shortcut {
	return Shortcut{Name: name, Action: action}
}

func (c *Config) RegisterShortcut(key string, shortcut Shortcut) {
	c.Shortcuts[key] = shortcut
}

func (c *Config) FindAndInvoke(key string, t any) bool {
	value, ok := c.Shortcuts[key]
	if ok {
		value.Action(t)
		return true
	}
	return false
}
