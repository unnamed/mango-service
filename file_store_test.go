package main

import "testing"

func TestCleanPath(t *testing.T) {
	CheckCleanPath(t, "data", "test.txt", true)
	CheckCleanPath(t, "data", "not/valid/test.txt", false)
	CheckCleanPath(t, "data", "../../../also-not-valid.txt", false)
	CheckCleanPath(t, "data", "validthing", true)
	CheckCleanPath(t, "data", "123valid", true)
	CheckCleanPath(t, "data", "./valid", false)
}

func CheckCleanPath(t *testing.T, directory string, name string, shouldBeValid bool) {
	path, err := cleanPath(directory, name)
	if shouldBeValid {
		if err != nil {
			t.Fatalf("Name '%s', in directory '%s', should be valid: %v", name, directory, err)
		}
	} else {
		if err == nil {
			t.Fatalf("Name '%s', in directory '%s', should not be valid, output was: '%s'", name, directory, path)
		}
	}
}