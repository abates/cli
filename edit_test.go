package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestEdit(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}
	os.Setenv("EDITOR", "")
	_, err := Edit([]byte{})
	if err != ErrNoEditor {
		t.Errorf("Wanted %v got %v", ErrNoEditor, err)
	}

	want := "this is some output"
	os.Setenv("EDITOR", os.Args[0])
	os.Setenv("TEST_OUTPUT", want)
	editCmd.Args = []string{"-run=TestHelperProcess", "--"}
	editCmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	gotBytes, err := Edit([]byte{})
	if err == nil {
		got := string(gotBytes)
		if want != got {
			t.Errorf("Wanted %q got %q", want, got)
		}
	} else {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No file\n")
		os.Exit(1)
	}

	err := ioutil.WriteFile(args[0], []byte(os.Getenv("TEST_OUTPUT")), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
