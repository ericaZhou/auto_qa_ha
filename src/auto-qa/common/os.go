package common

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func Cmd(str string) (output string, retCode int, err error) {
	var cmd *exec.Cmd
	if "windows" == runtime.GOOS {
		filename := fmt.Sprintf("command_%v.bat", rand.Int())
		defer os.Remove(filename)
		batFile := fmt.Sprintf("@echo off\n%v", str)
		if err := ioutil.WriteFile(filename, []byte(batFile), 0644); nil != err {
			return "", 1, err
		}
		cmd = exec.Command("cmd", "/C", filename)
	} else {
		cmd = exec.Command("sh", "-c", str)
	}
	o, e := cmd.CombinedOutput()
	if nil != e {
		if e2, ok := e.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				return string(o), int(s.ExitStatus()), nil
			}
		} else {
			return "", 0, e
		}
	}
	return string(o), 0, nil
}
