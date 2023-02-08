package logs

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/text"
	"strings"
	"todos/logs/prefixes"
)

func LogWarning(data ...string) {
	fmt.Print(prefixes.Pref.Err + text.FgYellow.Sprint(strings.Join(data, " ")))
}
func LogSuccess(data ...string) {
	fmt.Print(prefixes.Pref.Def + text.FgGreen.Sprint(strings.Join(data, " ")))
}
func Deb(data ...string) {
	fmt.Println(text.BgYellow.Sprint(text.FgBlack.Sprint(strings.Join(data, " "))))
}
func LogError(err error) {
	if err != nil {
		panic(err)
	}
	return
}
func NotEnoughArgs() {
	LogWarning("Not enough arguments, type `help` for help\n")
}