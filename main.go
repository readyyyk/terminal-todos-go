package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/ttacon/chalk"
	"golang.org/x/exp/slices"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var debStyle = chalk.Black.NewStyle().WithBackground(chalk.Yellow)

func deb(data string) {
	fmt.Println(debStyle.Style(data))
}
func logError(err error) {
	if err != nil {
		panic(err)
	}
	return
}
func getInput() []string {
	reader := bufio.NewScanner(os.Stdin)
	reader.Scan()
	logError(reader.Err())

	return strings.Split(reader.Text(), " ")
}

func executeDoskey() {
	localPath := "C:/bat"

	doskeyFile := file{
		path:         localPath + "/macros.doskey",
		defaultValue: "",
	}

	err := os.Mkdir(localPath, 0644)

	cmd := exec.Command("doskey", "todos="+path+"/main.exe")
	err = cmd.Run()
	//logError(err)

	doskeyFile.rewrite("todos=" + path + "/main.exe")

	//reg delete "HKEY_CURRENT_USER\Software\Microsoft\Command Processor" /v Autorun
	cmd = exec.Command("reg", "delete", `"HKCU\Software\Microsoft\Command Processor"`, "/v", "Autorun")
	err = cmd.Run()
	logError(err)

	cmd = exec.Command("reg", "add", `"HKCU\Software\Microsoft\Command Processor"`, "/v", "Autorun", "/d", `"doskey /macrofile=`+doskeyFile.path+`"`, "/f")
	err = cmd.Run()

	logError(err)
	return
}

type file struct {
	path         string
	defaultValue string
	// create()
	// read()
	// rewrite(value string)
}

func (r *file) create() (created bool) {
	if _, err := os.Stat(r.path); errors.Is(err, os.ErrNotExist) {
		file_, err := os.Create(r.path)
		logError(err)
		file_, err = os.Open(r.path)
		logError(err)
		err = os.WriteFile(r.path, []byte(r.defaultValue), 0644)
		logError(err)
		err = file_.Close()
		logError(err)
		return true
	}
	return false
}
func (r *file) read() (data []byte) {
	if r.create() {
		return []byte(r.defaultValue)
	}
	file_, err := os.Open(r.path)
	logError(err)
	data, err = io.ReadAll(file_)
	logError(err)
	err = file_.Close()
	logError(err)
	return data
}
func (r *file) rewrite(value string) {
	if r.create() {
		r.rewrite(value)
	}

	file_, err := os.Open(r.path)
	logError(err)
	err = os.WriteFile(r.path, []byte(value), 0644)
	logError(err)
	err = file_.Close()
	logError(err)
}

var dataFile file
var settingsFile file

const dateTimeFormat = "02.01_15:04"

var todoStates []string
var helpData table.Writer

func validateDate(value string) (isBefore bool, diff time.Duration, customError string) {
	utcDiff := time.Now().Hour() - time.Now().UTC().Hour()

	date, err := time.Parse(dateTimeFormat, value)
	if err != nil {
		//logError(err)
		return false, time.Duration(0), "Wrong datetime format (dd.MM_hh:mm)"
	}
	date = date.AddDate(time.Now().Year(), 0, 0)
	date = date.Add(-time.Duration(utcDiff) * time.Hour)

	isBefore = date.Before(time.Now())
	diff = date.Sub(time.Now())

	return isBefore, diff, ""
}

var path = ""

type prefixes struct {
	err      string
	def      string
	inp      string
	reset    string
	diff     string
	selector string

	yellow string
	green  string
	grey   string
}

var pref = prefixes{
	err:      " -x-> ",
	def:      " ---> ",
	inp:      "-> ",
	reset:    "",
	diff:     "",
	selector: "",
	// colors(enable bool)
}

func (r *prefixes) colors(enable bool) {
	if enable {
		pref = prefixes{
			err:      chalk.Yellow.String() + " -x-> ",
			def:      chalk.Green.String() + " ---> ",
			inp:      chalk.Magenta.String() + "-> ",
			reset:    chalk.Reset.String(),
			diff:     chalk.Cyan.String(),
			selector: chalk.Blue.String(),
		}
		fmt.Println(pref.def + "Colors enabled")
	} else {
		fmt.Print(chalk.Reset)
		pref = prefixes{
			err:      " -x-> ",
			def:      " ---> ",
			inp:      "-> ",
			reset:    "",
			diff:     "",
			selector: "",
		}
		fmt.Println(pref.def + "Colors disabled")
	}
}

type settings struct {
	Colors string
}

var settingsData settings

type todo struct {
	ID        int
	Title     string
	Text      string
	State     string
	Startdate string
	Deadline  string
	// delete() bool // todo
	// edit(field string, value string)
	// getFields() []string
}

func (r todo) edit(field string, value string) (newTodo todo, oldValue string, customError string) {

	temp := reflect.ValueOf(&r).Elem().FieldByName(field)
	if !temp.CanSet() {
		return r, "", "filed can't be assigned"
	}
	oldValue = temp.String()
	temp.SetString(value)
	//deb(temp.String())
	//data, _ := json.MarshalIndent(r, "", " ")
	//deb(string(data))

	return r, oldValue, ""
}

type todoArray struct {
	data []todo
	// get()
	// add(newTodo)
	// delete(id int) bool todo: move method to todo class
	// list()
	// drop()
}

var todos = todoArray{[]todo{}}

func (r *todoArray) get(type_ string) {
	switch type_ {
	case "json":
		dataBytes := dataFile.read()
		err := json.Unmarshal(dataBytes, &todos.data)
		logError(err)
		fmt.Println(pref.def+"Successfully read data file", pref.reset)
		return
	default:
		logError(errors.New("wrong get query type"))
		return
	}
}
func (r *todoArray) list() {
	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{"ID", "Title", "Text", "State", "Startdate", "Deadline"})
	for _, el := range r.data {
		//currentState := text.Colors{text.BgCyan, text.FgBlack}
		tbl.AppendRow(table.Row{el.ID, el.Title, el.Text, el.State, el.Startdate, el.Deadline})
		tbl.AppendSeparator()
	}
	tbl.SetCaption("github.com/readyyyk/terminal-todos")
	tbl.SetStyle(table.StyleBold)
	tbl.SetOutputMirror(os.Stdin)
	fmt.Println(tbl.Render())
}
func (r *todoArray) add(newTodo todo) {
	r.data = append(r.data, newTodo)
	data, err := json.MarshalIndent(r.data, "", "\t")
	logError(err)
	dataFile.rewrite(string(data))
}
func (r *todoArray) delete(id int) (found bool) {
	for i, el := range r.data {
		if el.ID == id {
			r.data = append(r.data[:i], r.data[i+1:]...)
			data, err := json.MarshalIndent(r.data, "", "\t")
			logError(err)
			dataFile.rewrite(string(data))
			return true
		}
	}
	return false
}

func doRequest(query []string) {
	//fmt.Println(debStyle.Style(strings.Join(query, ",")), pref.reset)
	switch query[0] {
	case "help":
		fmt.Println(helpData.Render())
	case "exit":
		fmt.Print(pref.reset)
		os.Exit(0)
	case "path":
		fmt.Println(pref.def+path, pref.reset)
	case "command":
		executeDoskey()
	case "colors":
		if len(query) < 2 {
			fmt.Println(pref.err+`Not enough arguments, type "help" for help`, pref.reset)
			break
		}
		if query[1] == "1" || query[1] == "en" || query[1] == "enable" {
			pref.colors(true)
			settingsFile.rewrite(`{"colors":"1"}`)
		} else if query[1] == "0" || query[1] == "da" || query[1] == "disable" {
			pref.colors(false)
			settingsFile.rewrite(`{"colors":"0"}`)
		} else {
			fmt.Println(pref.err+"Wrong query args", pref.reset)
		}
	case "ls":
		todos.list()
	case "list":
		todos.list()
	case "add":
		if len(query) < 4 {
			fmt.Println(pref.err+`Not enough arguments, type "help" for help`, pref.reset)
			break
		}
		tempId := 0
		if len(todos.data) != 0 {
			tempId = todos.data[len(todos.data)-1].ID + 1
		}

		newTodo := todo{
			ID:        tempId,
			Title:     strings.ReplaceAll(query[1], "_", " "),
			Text:      strings.ReplaceAll(query[2], "_", " "),
			State:     "passive",
			Startdate: time.Now().Format(dateTimeFormat),
		}

		if len(query) == 5 && query[4] == "t" {
			dur, err := time.ParseDuration(query[3])
			logError(err)
			newTodo.Deadline = time.Now().Add(dur).Format(dateTimeFormat)
			todos.add(newTodo)
			fmt.Println(pref.def+"Successfully added todo", pref.reset)
			break
		}
		if isBefore, _, customError := validateDate(query[3]); len(customError) > 0 {
			fmt.Println(pref.err+customError, pref.reset)
			break
		} else if isBefore {
			fmt.Println(pref.err+"Date is before now", pref.reset)
			break
		}
		newTodo.Deadline = query[3]
		todos.add(newTodo)
		fmt.Println(pref.def+"Successfully added todo", pref.reset)
	case "delete":
		if len(query) < 2 {
			fmt.Println(pref.err+`Not enough arguments, type "help" for help`, pref.reset)
			break
		}
		tempId, err := strconv.Atoi(query[1])
		if err != nil {
			fmt.Println(pref.err+"Wrong input, input a number", pref.reset)
			break
		}
		if todos.delete(tempId) {
			fmt.Println(pref.def+"Successfully deleted todo", pref.reset)
			break
		}
		fmt.Println(pref.err+"Can't find todo with", pref.selector+"{id} "+query[1], pref.reset)
	case "edit": // id field value
		if len(query) < 4 {
			fmt.Println(pref.err+`Not enough arguments, type "help" for help`, pref.reset)
			break
		}
		if tempId, err := strconv.Atoi(query[1]); err != nil {
			fmt.Println(pref.err+"Wrong input, input a valid", pref.selector+"ID", pref.reset)
			break
		} else if query[2] == "ID" || query[2] == "Startdate" {
			fmt.Println(pref.err+"Can't edit this field", pref.reset)
			break
		} else if isBefore, _, err := validateDate(query[3]); len(err) > 0 && query[2] == "Deadline" {
			fmt.Println(pref.err+"Error parsing date", pref.reset)
			break
		} else if isBefore && query[2] == "Deadline" {
			fmt.Println(pref.err+"Date is before now", pref.reset)
			break
		} else if validState := slices.IndexFunc(todoStates, func(r string) bool { return r == query[3] }); validState == -1 && query[2] == "State" {
			fmt.Println(pref.err+"Wrong input, input valid todo state", pref.reset)
			break
		} else {
			for i, el := range todos.data {
				if el.ID == tempId {
					newTodo, oldValue, customError := el.edit(query[2], query[3])
					if len(customError) > 0 {
						fmt.Println(pref.err+err, pref.selector+query[2], pref.reset)
						break
					}
					todos.data[i] = newTodo
					tempData, err := json.MarshalIndent(todos.data, "", "\t")
					logError(err)
					dataFile.rewrite(string(tempData))
					fmt.Println(pref.def+"Successfully edited todo with", pref.selector+"{ID} "+query[1], pref.diff+oldValue+" -> "+query[3], pref.reset)
					break
				} else if i == len(todos.data)-1 {
					fmt.Println(pref.err+"Can't find todo with", pref.selector+"{id} "+query[1], pref.reset)
				}
			}
		}
	default:
		fmt.Println(pref.err+"Wrong query", pref.reset)
	}
}

func init() {
	ex, err := os.Executable()
	logError(err)
	todoStates = []string{"passive", "in progress", "important", "done"}
	path = strings.ReplaceAll(filepath.Dir(ex), `\`, `/`)

	//fmt.Println(debStyle.Style(path))

	dataFile = file{
		path:         path + "/data.json",
		defaultValue: "[\n]",
	}
	settingsFile = file{
		path:         path + "/settings.json",
		defaultValue: `{"colors":"0"}`,
	}
	helpData = table.NewWriter()
	helpData.AppendRows([]table.Row{
		{"exit", "", ""},
		{"help", "", "prints help"},
		{"command", "", "program can be executed from any directory using `todos`"},
		{"colors", "1|0|enable|disable", "using to enable or disable color usage in program"},
		{"ls|list", "", "list all stored todos"},
		{"add", "{Title} {Text} {Deadline} (t)", "adds new todo, in case you enter duration {_}h{_}m type `t` in the end"},
		{"delete", "{ID}", "deletes todo"},
		{"edit", "{ID} {Field} {Value}", "edits todo"},
	})
	helpData.SetCaption("datetime format is: dd.MM_hh:mm (d - day, M - month, h - hour, m - minute)")
	helpData.SetStyle(table.StyleLight)
	helpData.Style().Options.SeparateRows = true
	helpData.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight},
	})
	helpData.SetStyle(table.StyleLight)
	helpData.Style().Options.DrawBorder = false
	helpData.Style().Options.SeparateColumns = true
	helpData.Style().Options.SeparateRows = true

	//helpData.SetColumnConfigs([]table.ColumnConfig{
	//	{Number: 2, AutoMerge: true},
	//})
	//helpData.SetOutputMirror(os.Stdin)

	logError(json.Unmarshal(settingsFile.read(), &settingsData))
	doRequest([]string{"colors", settingsData.Colors})

	todos.get("json")
}
func main() {
	defer fmt.Println(chalk.Reset)
	for {
		fmt.Print("\n" + pref.inp)
		query := getInput()

		doRequest(query)
	}
}
