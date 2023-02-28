package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/readyyyk/terminal-todos-go/pkg/files"
	"github.com/readyyyk/terminal-todos-go/pkg/logs"
	todos "github.com/readyyyk/terminal-todos-go/pkg/todoClasses"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/exp/slices"
)

func GetInput() []string {
	reader := bufio.NewScanner(os.Stdin)
	reader.Scan()
	logs.LogError(reader.Err())

	data := strings.Split(reader.Text(), " ")
	connect := false
	for i := 0; i < len(data); i++ {
		el := data[i]

		if re, _ := regexp.Compile("\".+\""); re.MatchString(el) {
			data[i] = strings.ReplaceAll(el, `"`, "")
			continue
		}

		if connect {
			data[i-1] = data[i-1] + " " + data[i]
			data[i-1] = strings.ReplaceAll(data[i-1], `"`, "")
			data = append(data[:i], data[i+1:]...)
			i--
		}
		if strings.Index(el, `"`) == 0 {
			connect = true
		} else if strings.Index(el, `"`) == len(el)-1 {
			connect = false
		}
	}

	//logs.Deb(strings.Join(data, "\n"))

	return data
}
func approve() bool {
	logs.LogWarning("Are you sure? (`y` to continue), type help for more information\n")
	inputData := strings.ToLower(GetInput()[0])
	return inputData == "y" || inputData == "yes"
}

func executeDoskey() {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		err := exec.Command("bash", "-c", "echo \"\" >> ~/.bashrc; echo 'export PATH=\"$PATH:"+path+"\"' >> ~/.bashrc").Run()
		logs.LogError(err)

		logs.LogSuccess("Relaunch console to update Path variables\n\tThen you are able to run app with `todos`\n\tEnjoy)\n")
		os.Exit(0)
		return
	}
	cmd := exec.Command("setx", "path", os.Getenv("path")+";"+path)
	var stderrText bytes.Buffer
	cmd.Stderr = &stderrText
	err := cmd.Run()
	if err != nil && strings.Contains(stderrText.String(), "denied") {
		logs.LogWarning("App has no access to path variable, so, run next command and relaunch console for using command `todos.exe`\n(!!! Run with admin rights !!!)\n")
		fmt.Println(`setx path "%path%;` + path + `"`)
		return
	}
	logs.LogError(err)

	logs.LogSuccess("Relaunch console to update Path variables\n\tThen you are able to run app with `todos`\n\tEnjoy)\n")
	os.Exit(0)
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
	if time.Now().Day() != time.Now().UTC().Day() {
		date = date.Add(-time.Hour * time.Duration(24))
	}

	isBefore = date.Before(time.Now())
	diff = date.Sub(time.Now())

	return isBefore, diff, ""
}

var path = ""

type settings struct {
	Colors    bool
	AutoSort  bool
	SortValue string
	AutoClear bool
}

var settingsData settings

func cls() {
	fmt.Print("\033[H\033[2J")
	/*if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		logs.LogError(cmd.Run())
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Run()
		cmd.Stdout = os.Stdout
		// logs.LogError(err)
	}*/
}
func doRequest(query []string) {
	if settingsData.AutoClear {
		cls()
	}

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
	case "autosort":
		settingsData.AutoSort = !settingsData.AutoSort
		data, err := json.MarshalIndent(settingsData, "", "\t")
		logs.LogError(err)
		err = settingsFile.Rewrite(string(data))
		logs.LogError(err)
		logs.LogSuccess("Autosort successfully set\n")
	case "autoclear":
		settingsData.AutoClear = !settingsData.AutoClear
		data, err := json.MarshalIndent(settingsData, "", "\t")
		logs.LogError(err)
		err = settingsFile.Rewrite(string(data))
		logs.LogError(err)
		logs.LogSuccess("Autoclear successfully set\n")
	case "colors":
		settingsData.Colors = !settingsData.Colors
		data, err := json.MarshalIndent(settingsData, "", "\t")
		logs.LogError(err)
		err = settingsFile.Rewrite(string(data))
		logs.LogError(err)

		logs.Pref.Colors(settingsData.Colors)
		if settingsData.Colors {
			logs.LogSuccess("Colors enabled\n")
		} else {
			logs.LogSuccess("Colors disabled\n")
		}
	//case "cls":
	case "ls":
		Todos.List(validateDate)
	case "list":
		Todos.List(validateDate)
	case "clear":
		if !approve() {
			logs.LogSuccess("Canceled\n")
			break
		}
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

		settingsData.SortValue = query[1]
		data, err = json.MarshalIndent(settingsData, "", "\t")
		logs.LogError(err)
		err = settingsFile.Rewrite(string(data))
		logs.LogError(err)

		logs.LogSuccess("Todos sorted with", logs.Pref.Selector(query[1]), "\n")
	case "add":
		if len(query) < 4 {
			logs.NotEnoughArgs()
			break
		}
		tempId := 0
		if len(Todos.Data) != 0 {
			tempId = Todos.Data[0].ID + 1
			for _, el := range Todos.Data {
				tempId = int(math.Max(float64(el.ID)+1, float64(tempId)))
			}
		}

		newTodo := todos.Todo{
			ID:        tempId,
			Title:     query[1],
			Text:      query[2],
			State:     "passive",
			Startdate: time.Now().Format(DateTimeFormat),
		}

		if len(query) == 5 && query[4] == "t" {
			dur, err := time.ParseDuration(query[3])
			logs.LogError(err)
			newTodo.Deadline = time.Now().Add(dur).Format(DateTimeFormat)
			Todos.Add(newTodo)
			logs.LogSuccess("Successfully added Todo\n")

			data, err := json.MarshalIndent(Todos.Data, "", "\t")
			logs.LogError(err)
			err = dataFile.Rewrite(string(data))
			logs.LogError(err)

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
				fmt.Print(logs.Pref.Selector("id="+strconv.Itoa(ids[i])), "\n")
			}
		}
	case "edit":
		if len(query) < 4 {
			logs.NotEnoughArgs()
			break
		}

		if query[2] == "State" {
			query[3] = query[3]
		}

		if tempId, err := strconv.Atoi(query[1]); err != nil {
			logs.LogWarning("Wrong input, input a valid ")
			fmt.Print(logs.Pref.Selector("ID"), "\n")
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
						fmt.Print(" "+logs.Pref.Selector(query[2]), "\n")
						break
					}
					Todos.Data[i] = newTodo

					data, err := json.MarshalIndent(Todos.Data, "", "\t")
					logs.LogError(err)
					err = dataFile.Rewrite(string(data))
					logs.LogError(err)

					logs.LogSuccess("Successfully edited Todo with ")
					fmt.Print(logs.Pref.Selector("{ID}="+query[1]), " ", logs.Pref.Diff(oldValue, query[3]), "\n")
					break
				} else if i == len(Todos.Data)-1 {
					logs.LogWarning("Can't find Todo with ")
					fmt.Print(logs.Pref.Selector("{id}="+query[1]), "\n")
				}
			}
		}
	default:
		logs.LogWarning("Wrong query\n")
	}

	if settingsData.AutoSort {
		err := Todos.Sort(settingsData.SortValue, DateTimeFormat)
		logs.LogError(err)
	}
}

func init() {
	ex, err := os.Executable()
	logs.LogError(err)
	TodoStates = []string{"passive", "in progress", "important", "done"}
	path = filepath.Dir(ex)

	dataFile = files.File{
		Path:         path + "/Data.json",
		DefaultValue: "[\n]",
	}
	settingsFile = files.File{
		Path:         path + "/settings.json",
		DefaultValue: `{"Colors": false, "AutoSort": false, "SortValue":"ID", "AutoClear": false}`,
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
		{"command", "", "program can be executed from any directory using `todos`"},

		{"colors", "", "toggle enable or disable color usage"},
		{"autosort", "", "toggle enable or disable automatic sorting with field in last `sort`"},
		{"autoclear", "", "toggle enable or disable automatic clearing screen"},

		{"ls|list", "", "list all stored Todos"},
		{"clear", "", "deletes all todos with `done` State"},
		{"sort", "{Field}", "sorts todos array with the Field"},
		{"add", "{Title} {Text} {Deadline} (t)", "adds new Todo, in case you enter duration {_}h{_}m type `t` in the end"},
		{"delete", "{ID_1 ID_2 ID_3...}", "deletes Todo"},
		{"edit", "{ID} {Field} {Value}", "edits Todo"},
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

	logs.Pref.Colors(settingsData.Colors)

	Todos.Get("json")
	logs.LogSuccess("Successfully read Data file\n\n")
	logs.LogSuccess("Type `help` for help\n")
}
func main() {
	for {
		fmt.Print("\n" + logs.Pref.Inp)
		query := GetInput()

		doRequest(query)
	}
}
