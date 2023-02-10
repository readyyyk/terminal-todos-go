package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"todos/files"
	"todos/logs"
	"todos/logs/prefixes"
	"todos/todoClasses"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/exp/slices"
)

func getInput() []string {
	reader := bufio.NewScanner(os.Stdin)
	reader.Scan()
	logs.LogError(reader.Err())

	return strings.Split(reader.Text(), " ")
}

func executeDoskey() {
	//reg add "HKCU\Enviroment" /v todos /d "d:/.prog/_go/todos/todos.exe" /f
	cmd := exec.Command("setx", "todos", `"`+path+`/todos.exe"`, "/f")
	err := cmd.Run()
	logs.LogError(err)

	cmd = exec.Command("reg", "add", `HKCU\Enviroment`, "/v", "todos", "/d", `"`+path+`/todos.exe"`)
	err = cmd.Run()
	logs.LogError(err)
	return
}

var dataFile files.File
var settingsFile files.File

const DateTimeFormat = "02.01_15:04"

var TodoStates []string
var helpData table.Writer

var Todos todos.TodoArray

func validateDate(value string) (isBefore bool, diff time.Duration, customError string) {
	utcDiff := time.Now().Hour() - time.Now().UTC().Hour()

	date, err := time.Parse(DateTimeFormat, value)
	if err != nil {
		//logs.LogError(err)
		return false, time.Duration(0), "Wrong datetime format (dd.MM_hh:mm)"
	}
	date = date.AddDate(time.Now().Year(), 0, 0)
	date = date.Add(-time.Duration(utcDiff) * time.Hour)

	isBefore = date.Before(time.Now())
	diff = date.Sub(time.Now())

	return isBefore, diff, ""
}

var path = ""

type settings struct {
	Colors string
}

var settingsData settings

func doRequest(query []string) {
	switch query[0] {
	case "help":
		fmt.Println("\n" + helpData.Render() + "\n")
		fmt.Println("datetime format is: dd.MM_hh:mm (d - day, M - month, h - hour, m - minute)")
		fmt.Println("duration format is: {_}h{_}m (for example 12h30m, or 1h1m, but not 1d12h)")
	case "exit":
		os.Exit(0)
	case "path":
		logs.LogSuccess(path, "\n")
	case "command":
		executeDoskey()
		logs.LogSuccess("Command `todos` set")
	case "colors":
		if len(query) < 2 {
			logs.NotEnoughArgs()
			break
		}
		if query[1] == "1" || query[1] == "enable" {
			prefixes.Pref.Colors(true)
			logs.LogSuccess("Colors enabled\n")
			err := settingsFile.Rewrite(`{"colors":"1"}`)
			logs.LogError(err)
		} else if query[1] == "0" || query[1] == "disable" {
			prefixes.Pref.Colors(false)
			logs.LogSuccess("Colors disabled\n")
			err := settingsFile.Rewrite(`{"colors":"0"}`)
			logs.LogError(err)
		} else {
			logs.LogWarning("Wrong query args")
		}
	case "ls":
		Todos.List(validateDate)
	case "list":
		Todos.List(validateDate)
	case "clear":
		Todos.Clear()

		data, err := json.MarshalIndent(Todos.Data, "", "\t")
		logs.LogError(err)
		err = dataFile.Rewrite(string(data))
		logs.LogError(err)

		logs.LogSuccess("Todos with `done` state deleted\n")
	case "sort":
		err := Todos.Sort(query[1], DateTimeFormat)

		data, err := json.MarshalIndent(Todos.Data, "", "\t")
		logs.LogError(err)
		err = dataFile.Rewrite(string(data))
		logs.LogError(err)

		logs.LogError(err)
		logs.LogSuccess("Todos with `done` state deleted\n")
	case "add":
		if len(query) < 4 {
			logs.NotEnoughArgs()
			break
		}
		tempId := 0
		if len(Todos.Data) != 0 {
			tempId = Todos.Data[len(Todos.Data)-1].ID + 1
		}

		newTodo := todos.Todo{
			ID:        tempId,
			Title:     strings.ReplaceAll(query[1], "_", " "),
			Text:      strings.ReplaceAll(query[2], "_", " "),
			State:     "passive",
			Startdate: time.Now().Format(DateTimeFormat),
		}

		if len(query) == 5 && query[4] == "t" {
			dur, err := time.ParseDuration(query[3])
			logs.LogError(err)
			newTodo.Deadline = time.Now().Add(dur).Format(DateTimeFormat)
			Todos.Add(newTodo)
			logs.LogSuccess("Successfully added Todo\n")
			break
		}
		if isBefore, _, customError := validateDate(query[3]); len(customError) > 0 {
			logs.LogWarning(customError)
			break
		} else if isBefore {
			logs.LogWarning("Date is before now")
			break
		}
		newTodo.Deadline = query[3]
		Todos.Add(newTodo)
		data, err := json.MarshalIndent(Todos.Data, "", "\t")
		logs.LogError(err)
		err = dataFile.Rewrite(string(data))
		logs.LogError(err)
		logs.LogSuccess("Successfully added Todo\n")
	case "delete":
		if len(query) < 2 {
			logs.NotEnoughArgs()
			break
		}
		var idArray []int
		for i, el := range query {
			if i > 0 {
				tempId, err := strconv.Atoi(el)
				if err != nil {
					logs.LogWarning("Wrong input, input a number\n")
					continue
				}
				idArray = append(idArray, tempId)
			}
		}
		foundArray, ids := Todos.Delete(idArray)

		data, err := json.MarshalIndent(Todos.Data, "", "\t")
		logs.LogError(err)
		err = dataFile.Rewrite(string(data))
		logs.LogError(err)

		for i, found := range foundArray {
			if found {
				logs.LogSuccess("Successfully deleted Todo\n")
			} else {
				logs.LogWarning("Can't find Todo with ")
				fmt.Print(prefixes.Pref.Selector("id="+strconv.Itoa(ids[i])), "\n")
			}
		}
	case "edit":
		if len(query) < 4 {
			logs.NotEnoughArgs()
			break
		}

		if query[2] == "State" {
			query[3] = strings.ReplaceAll(query[3], "_", " ")
		}

		if tempId, err := strconv.Atoi(query[1]); err != nil {
			logs.LogWarning("Wrong input, input a valid ")
			fmt.Print(prefixes.Pref.Selector("ID"), "\n")
			break
		} else if query[2] == "ID" || query[2] == "Startdate" {
			logs.LogWarning("Can't Edit this field\n")
			break
		} else if isBefore, _, err := validateDate(query[3]); len(err) > 0 && query[2] == "Deadline" {
			logs.LogWarning("Error parsing date\n")
			break
		} else if isBefore && query[2] == "Deadline" {
			logs.LogWarning("Date is before now\n")
			break
		} else if validState := slices.IndexFunc(TodoStates, func(r string) bool { return r == query[3] }); validState == -1 && query[2] == "State" {
			logs.LogWarning("Wrong input, input valid Todo state\n")
			break
		} else {
			for i, el := range Todos.Data {
				if el.ID == tempId {
					newTodo, oldValue, customError := el.Edit(query[2], query[3])
					if len(customError) > 0 {
						logs.LogWarning(err)
						fmt.Print(" "+prefixes.Pref.Selector(query[2]), "\n")
						break
					}
					Todos.Data[i] = newTodo

					data, err := json.MarshalIndent(Todos.Data, "", "\t")
					logs.LogError(err)
					err = dataFile.Rewrite(string(data))
					logs.LogError(err)

					logs.LogSuccess("Successfully edited Todo with ")
					fmt.Print(prefixes.Pref.Selector("{ID}="+query[1]), " ", prefixes.Pref.Diff(oldValue, query[3]), "\n")
					break
				} else if i == len(Todos.Data)-1 {
					logs.LogWarning("Can't find Todo with ")
					fmt.Print(prefixes.Pref.Selector("{id}="+query[1]), "\n")
				}
			}
		}
	default:
		logs.LogWarning("Wrong query\n")
	}
}

func init() {
	ex, err := os.Executable()
	logs.LogError(err)
	TodoStates = []string{"passive", "in progress", "important", "done"}
	path = strings.ReplaceAll(filepath.Dir(ex), `\`, `/`)

	dataFile = files.File{
		Path:         path + "/Data.json",
		DefaultValue: "[\n]",
	}
	settingsFile = files.File{
		Path:         path + "/settings.json",
		DefaultValue: `{"colors":"0"}`,
	}

	Todos = todos.TodoArray{
		Data:     []todos.Todo{},
		DataFile: dataFile,
		Origin:   "",
	}

	helpData = table.NewWriter()
	helpData.AppendRows([]table.Row{
		{"exit", "", ""},
		{"help", "", "prints help"},
		{"command", "", "program can be executed from any directory using `Todos`"},
		{"colors", "1|0|enable|disable", "using to enable or disable color usage in program"},
		{"ls|list", "", "list all stored Todos"},
		{"add", "{Title} {Text} {Deadline} (t)", "adds new Todo, in case you enter duration {_}h{_}m type `t` in the end"},
		{"delete", "{ID_1 ID_2 ID_3...}", "deletes Todo"},
		{"Edit", "{ID} {Field} {Value}", "edits Todo"},
		{"Clear", "", "deletes all todos with `done` State"},
	})
	helpData.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight},
		{Number: 2, Align: text.AlignCenter},
	})
	helpData.SetStyle(table.StyleLight)
	helpData.Style().Options.SeparateRows = true
	helpData.Style().Options.DrawBorder = false

	tempData, err := settingsFile.Read()
	logs.LogError(err)
	logs.LogError(json.Unmarshal(tempData, &settingsData))
	doRequest([]string{"colors", settingsData.Colors})

	Todos.Get("json")
	logs.LogSuccess("Successfully read Data file\n")
}
func main() {
	for {
		fmt.Print("\n" + prefixes.Pref.Inp)
		query := getInput()

		doRequest(query)
	}
}
