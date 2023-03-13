package todos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/readyyyk/terminal-todos-go/pkg/files"
	"github.com/readyyyk/terminal-todos-go/pkg/logs"
	"reflect"
	"sort"
	"time"
)

type Todo struct {
	ID        int
	Title     string
	Text      string
	State     string
	Startdate string
	Deadline  string
	// Edit(field string, value string)
	// getFields() []string
}

func (r Todo) Edit(field string, value string) (newTodo Todo, oldValue string, customError string) {
	temp := reflect.ValueOf(&r).Elem().FieldByName(field)
	if !temp.CanSet() {
		return r, "", "filed can't be assigned"
	}
	oldValue = temp.String()
	temp.SetString(value)

	return r, oldValue, ""
}

type TodoArray struct {
	Data     []Todo
	DataFile files.File
	Origin   string
	// Get()
	// Add(newTodo)
	// List()
	// Drop() todo
}

func (r *TodoArray) Get(type_ string) {
	switch type_ {
	case "json":
		dataBytes, err := r.DataFile.Read()
		logs.LogError(err)
		err = json.Unmarshal(dataBytes, &r.Data)
		logs.LogError(err)
		return
	default:
		logs.LogError(errors.New("wrong get query type"))
		return
	}
}
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02dh %02dm", h, m)
}

// List : refactor using function for `time left`
func (r *TodoArray) List(validateDate func(string) (bool, time.Duration, string)) {
	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{text.FgHiBlack.Sprint("ID"), "Title", "Text", "State", text.FgHiBlack.Sprint("time left"), text.FgHiBlack.Sprint("Startdate"), "Deadline"})
	for _, el := range r.Data {
		//currentState := text.Colors{text.BgCyan, text.FgBlack}

		//stateString := el.State
		switch el.State {
		case "passive":
			el.State = text.FgHiBlack.Sprint(el.State)
		case "in progress":
			el.State = text.FgCyan.Sprint(el.State)
		case "important":
			el.State = text.FgYellow.Sprint(el.State)
		case "done":
			el.State = text.FgGreen.Sprint(el.State)
		}

		_, timeLeft, _ := validateDate(el.Deadline)
		timeLeftString := formatDuration(timeLeft) //.Round(time.Minute).String()
		if timeLeft < time.Duration(0) {
			timeLeftString = text.FgRed.Sprint("time up")
			el.Deadline = text.FgRed.Sprint(el.Deadline)
		} else if timeLeft < time.Hour*3 {
			timeLeftString = text.FgYellow.Sprint(timeLeftString)
			el.Deadline = text.FgYellow.Sprint(el.Deadline)
		}

		tbl.AppendRow(table.Row{text.FgHiBlack.Sprint(el.ID), el.Title, el.Text, el.State, timeLeftString, el.Startdate, el.Deadline})
	}
	tbl.SetCaption("github.com/readyyyk/terminal-todos-go")
	tbl.SetStyle(table.StyleBold)
	tbl.Style().Format.Header = text.FormatDefault
	tbl.Style().Options.SeparateRows = true

	fmt.Println(tbl.Render())
}
func (r *TodoArray) Add(newTodo Todo) {
	r.Data = append(r.Data, newTodo)
}
func (r *TodoArray) Delete(id []int) (found []bool, ids []int) {
	//found = []bool{}
	for _, deleteAble := range id {
		for i, el := range r.Data {
			if el.ID == deleteAble {
				r.Data = append(r.Data[:i], r.Data[i+1:]...)
				found = append(found, true)
				ids = append(ids, deleteAble)
			}
		}
		if len(ids) == 0 || ids[len(ids)-1] != deleteAble {
			found = append(found, false)
			ids = append(ids, deleteAble)
		}
	}

	return found, ids
}
func (r *TodoArray) Clear() {
	var toDel []int
	for _, el := range r.Data {
		if el.State == "done" {
			toDel = append(toDel, el.ID)
		}
	}
	r.Delete(toDel)
}
func (r *TodoArray) Sort(field string, dateTimeFormat string) error {
	sort.Slice(r.Data, func(i, j int) bool {
		temp1 := reflect.ValueOf(&r.Data[i]).Elem().FieldByName(field)
		temp2 := reflect.ValueOf(&r.Data[j]).Elem().FieldByName(field)

		if temp1.CanInt() {
			return temp1.Int() < temp2.Int()
		} else {
			td1, err := time.Parse(dateTimeFormat, temp1.String())
			td2, err := time.Parse(dateTimeFormat, temp2.String())
			if err != nil {
				return temp1.String() < temp2.String()
			}
			return td1.Before(td2)
		}

		//logs.Deb(temp1, "\n", temp2, "\n")

	})
	return nil
}
