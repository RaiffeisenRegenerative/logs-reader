package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

func newCustomDelegate(keys *keyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	help := []key.Binding{
		keys.filterEmail,
		keys.filterNextjs,
		keys.clearOrigin,
		keys.filterDebug,
		keys.filterInfo,
		keys.filterWarn,
		keys.filterError,
		keys.filterFatal,
		keys.clearLevel,
		keys.quit,
	}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
