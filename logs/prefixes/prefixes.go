package prefixes

import (
	"github.com/jedib0t/go-pretty/v6/text"
)

type Prefixes struct {
	Err string
	Def string
	Inp string

	DiffColor     text.Color
	SelectorColor text.Color
	// Diff(oldValue, newValue string) string
	// Selector(field, value string) string
	// Colors(enable bool)
}

var Pref = Prefixes{
	Err: text.FgYellow.Sprint(" -x-> "),
	Def: text.FgGreen.Sprint(" ---> "),
	Inp: text.FgMagenta.Sprint(" -> "),

	DiffColor:     text.FgCyan,
	SelectorColor: text.FgBlue,
}

func (r *Prefixes) Colors(enable bool) {
	if enable {
		text.EnableColors()
	} else {
		text.DisableColors()
	}
}
func (r *Prefixes) Diff(oldValue, newValue string) string {
	return r.DiffColor.Sprintf("%s -> %s", oldValue, newValue)
}
func (r *Prefixes) Selector(data string) string {
	return r.SelectorColor.Sprint(data)
}
