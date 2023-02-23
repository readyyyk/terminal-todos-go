package logs

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

const (
	PrefErrorString   = " -x-> "
	PrefSuccessString = " ---> "
	PrefInputString   = " -> "
)

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
		Pref.Err = text.FgYellow.Sprint(PrefErrorString)
		Pref.Def = text.FgGreen.Sprint(PrefSuccessString)
		Pref.Inp = text.FgMagenta.Sprint(PrefInputString)
	} else {
		text.DisableColors()
		Pref.Err = PrefErrorString
		Pref.Def = PrefSuccessString
		Pref.Inp = PrefInputString
	}
}
func (r *Prefixes) Diff(oldValue, newValue string) string {
	return r.DiffColor.Sprintf("%s -> %s", oldValue, newValue)
}
func (r *Prefixes) Selector(data string) string {
	return r.SelectorColor.Sprint(data)
}
