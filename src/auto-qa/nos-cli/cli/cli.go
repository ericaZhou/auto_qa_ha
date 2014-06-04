package cli

import (
	"auto-qa/nos-cli/agent"
	"auto-qa/nos-cli/tcp"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
)

var remoteOs = "linux"

func InitRemoteOs(os string) {
	remoteOs = os
}

func Cmdf(ip, cmd string, args ...interface{}) (output string, exitCode int, err error) {
	gob.Register(agent.AgentRequest{})
	gob.Register(agent.AgentResponse{})
	request := agent.AgentRequest{}
	request.Op = "cmd"
	request.Command = fmt.Sprintf(cmd, args...)
	response, err := callRemote(ip, "5555", request)
	if nil != err {
		return "", 0, err
	}
	return response.Output, response.ExitCode, nil
}

func PushFile(ip, filePath string, targetFilePath string) error {
	file, err := ioutil.ReadFile(filePath)
	if nil != err {
		return err
	}
	request := agent.AgentRequest{}
	request.Op = "push_file"
	request.FilePath = targetFilePath
	request.FileSize = int64(len(file))

	err = tcp.NewTcp().SendMsg(ip, "5555", request, func(conn net.Conn) error {
		var response string
		err := gob.NewDecoder(conn).Decode(&response)
		if nil != err {
			return err
		}
		if "succeed" != response {
			return fmt.Errorf("push file got unexpected response %v", response)
		}
		_, err = conn.Write(file)
		if nil != err {
			return err
		}
		return nil
	})

	time.Sleep(2 * time.Second) //wait recv finish writing file

	return err
}

func PullFile(ip, filePath string, targetFilePath string) error {
	request := agent.AgentRequest{}
	request.Op = "pull_file"
	request.FilePath = filePath
	err := tcp.NewTcp().SendMsg(ip, "5555", request, func(conn net.Conn) error {
		var response agent.AgentResponse
		if err := gob.NewDecoder(conn).Decode(&response); nil != err {
			return err
		}
		if 0 != response.ExitCode {
			return fmt.Errorf("pull file got remote error : %v", response.Output)
		}
		fileSize := response.FileSize
		gob.NewEncoder(conn).Encode("start transfer")
		buf := make([]byte, 10000)
		f, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE, 0644)
		if nil != err {
			return err
		}
		defer f.Close()
		f.Seek(0, 2)
		for fileSize > 0 {
			n, err := conn.Read(buf)
			if nil != err {
				break
			}
			f.Write(buf[:n])
			fileSize -= int64(n)
		}
		return nil
	})
	return err
}

func FindLog(ip string, logPath string, logStartPos int64, expectedRegexp string) (found bool, newCheckpointPos int64, err error) {
	request := agent.AgentRequest{}
	request.Op = "find_log"
	request.FilePath = logPath
	request.LogStartPos = logStartPos
	request.LogExpectRegexp = expectedRegexp

	response, err := callRemote(ip, "5555", request)
	if nil != err {
		return false, 0, err
	}
	if -1 == response.LogCheckpointPos {
		return 0 == response.ExitCode, logStartPos, nil
	} else {
		return 0 == response.ExitCode, response.LogCheckpointPos, nil
	}
	return false, 0, nil
}

func FindLogPos(ip, logPath string) (newCheckpointPos int64, err error) {
	request := agent.AgentRequest{}
	request.Op = "find_log_pos"
	request.FilePath = logPath

	response, err := callRemote(ip, "5555", request)
	if nil != err {
		return 0, err
	}
	return response.LogCheckpointPos, nil
}

func callRemote(ip, port string, request agent.AgentRequest) (agent.AgentResponse, error) {
	var response agent.AgentResponse
	err := tcp.NewTcp().SendMsg(ip, port, request, func(conn net.Conn) error {
		err := gob.NewDecoder(conn).Decode(&response)
		if nil != err {
			return err
		}
		return nil
	})
	if nil != err {
		return agent.AgentResponse{}, fmt.Errorf("call remote err : %v\n", err)
	}
	return response, nil
}

func RmDir(ip, path string) (output string, exitCode int, err error) {
	if "windows" == remoteOs {
		Cmdf(ip, `c:\Unlocker.exe %v -S`, path)
		return Cmdf(ip, "rmdir /S /Q %v", path)
	}
	return Cmdf(ip, fmt.Sprintf("rm -r -f %v", path))
}

func Rm(ip, path string) (output string, exitCode int, err error) {
	if "windows" == remoteOs {
		Cmdf(ip, fmt.Sprintf(`c:\Unlocker.exe %v -S`, path))
		return Cmdf(ip, "del /F /Q %v", path)
	}
	return Cmdf(ip, fmt.Sprintf("rm -r -f %v", path))
}
