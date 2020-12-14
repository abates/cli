package cli

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
)

var ErrNoEditor = errors.New("No editor found in environment")

var editCmd = &exec.Cmd{}

func Edit(input []byte) (output []byte, err error) {
	editCmd.Path = os.Getenv("EDITOR")
	if editCmd.Path == "" {
		err = ErrNoEditor
	} else {
		editCmd.Path, err = exec.LookPath(editCmd.Path)
		if err == nil {
			var tmpfile *os.File
			tmpfile, err = ioutil.TempFile("", "")
			if err == nil {
				defer os.Remove(tmpfile.Name())
				tmpfile.Write(input)

				if err = tmpfile.Close(); err == nil {
					editCmd.Args = append(editCmd.Args, tmpfile.Name())
					editCmd.Stdin = os.Stdin
					editCmd.Stdout = os.Stdout
					editCmd.Stderr = os.Stderr
					editCmd.Start()
					err = editCmd.Wait()
					if err == nil {
						output, err = ioutil.ReadFile(tmpfile.Name())
					}
				}
			}
		}
	}
	return
}
