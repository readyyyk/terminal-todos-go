package files

import (
	"errors"
	"io"
	"os"
)

type File struct {
	Path         string
	DefaultValue string
	// create()
	// read()
	// rewrite(value string)
}

func (r *File) Create() (created bool, err error) {
	if _, err := os.Stat(r.Path); errors.Is(err, os.ErrNotExist) {
		file_, err := os.Create(r.Path)
		if err != nil {
			return false, err
		}
		file_, err = os.Open(r.Path)
		if err != nil {
			return false, err
		}
		err = os.WriteFile(r.Path, []byte(r.DefaultValue), 0644)
		if err != nil {
			return false, err
		}
		err = file_.Close()
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
func (r *File) Read() (data []byte, err error) {
	created, err := r.Create()
	if err != nil {
		return []byte{}, err
	}
	if created {
		return []byte(r.DefaultValue), nil
	}
	file_, err := os.Open(r.Path)
	if err != nil {
		return []byte{}, err
	}
	data, err = io.ReadAll(file_)
	if err != nil {
		return []byte{}, err
	}
	err = file_.Close()
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}
func (r *File) Rewrite(value string) (err error) {
	created, err := r.Create()
	if created {
		return r.Rewrite(value)
	}

	file_, err := os.Open(r.Path)
	if err != nil {
		return err
	}
	err = os.WriteFile(r.Path, []byte(value), 0644)
	if err != nil {
		return err
	}
	err = file_.Close()
	if err != nil {
		return err
	}
	return nil
}
