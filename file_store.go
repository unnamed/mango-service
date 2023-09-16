package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// Gets the clean path for the given file name, verifies that the
// given name is the name of a single file and does not contain
// anything like "..", ".", "/", for security reasons
func cleanPath(directory string, name string) (string, error) {
	path := filepath.Join(directory, name)
	path = filepath.Clean(path)

	if !strings.HasPrefix(path, directory) {
		// if path does not start with local directory,
		// it probably contained "..", and thus is invalid
		return "", fmt.Errorf("invalid file name '%s' (skipped prefix)", name)
	}

	_, filename := filepath.Split(path)
	if filename != name {
		// makes sure that 'name' did not contain file separators
		return "", fmt.Errorf("invalid file name '%s' (unexpected args)", name)
	}

	return path, nil
}

func (local *LocalFileStore) Exists(name string) bool {
	path, err := cleanPath(local.directory, name)
	if err != nil {
		// invalid file names don't exist.
		return false
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (local *LocalFileStore) Create(name string, data io.Reader) error {

	dir := local.directory
	path, err := cleanPath(dir, name)

	if err != nil {
		// invalid file name, don't continue
		return err
	}

	// create directory if it doesn't exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil
	}

	// fail to create already existent files
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
	path, err := cleanPath(dir, name)

	if err != nil {
		// invalid file name
		return nil, err
	}

	// check if file doesn't exist
	if _, err = os.Stat(path); os.IsNotExist(err) {
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
	path, err := cleanPath(local.directory, name)
	if err != nil {
		// invalid file name
		return false, err
	}
	if local.Exists(name) {
		return true, os.Remove(path)
	} else {
		return false, nil
	}
}
