package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStore interface {
	Exists(name string) bool

	Create(name string, data io.Reader) error

	Read(name string) (*os.File, error)

	List() ([]string, error)

	Delete(name string) (bool, error)
}

type LocalFileStore struct {
	directory string
}

func (local *LocalFileStore) Exists(name string) bool {
	path := filepath.Join(local.directory, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (local *LocalFileStore) Create(name string, data io.Reader) error {
	dir := local.directory

	// create directory if it doesn't exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil
	}

	// fail to create already existent files
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("can't create '%s', it already exists", name)
	} else if !os.IsNotExist(err) {
		return err
	}

	// create
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// copy data
	_, err = io.Copy(file, data)
	return err
}

func (local *LocalFileStore) Read(name string) (*os.File, error) {
	dir := local.directory
	path := filepath.Join(dir, name)

	// check if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file with name '%s' doesn't exist", name)
	}

	// just open if it exists
	return os.Open(path)
}

func (local *LocalFileStore) List() ([]string, error) {
	dir := local.directory
	var list []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// return empty
			return list, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			list = append(list, entry.Name())
		}
	}

	return list, nil
}

func (local *LocalFileStore) Delete(name string) (bool, error) {
	if local.Exists(name) {
		path := filepath.Join(local.directory, name)
		return true, os.Remove(path)
	} else {
		return false, nil
	}
}
