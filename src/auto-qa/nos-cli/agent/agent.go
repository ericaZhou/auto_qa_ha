package agent

import (
	"auto-qa/common"
	"auto-qa/nos-cli/tcp"
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
)

type AgentRequest struct {
	Op              string
	Command         string
	FilePath        string
	FileSize        int64
	LogStartPos     int64
	LogExpectRegexp string
}

type AgentResponse struct {
	Output           string
	ExitCode         int
	FileSize         int64
	LogCheckpointPos int64 //-1 means no change, otherwise cli should update log checkpoint since log file rotation
}

func StartAgentServer() {
	gob.Register(AgentRequest{})
	gob.Register(AgentResponse{})
	tcp.NewTcp().StartServer("0.0.0.0", "5555", func(conn net.Conn) {
		request := AgentRequest{}
		err := gob.NewDecoder(conn).Decode(&request)
		if nil != err {
			fmt.Printf("agent server error : gob error : %v", err)
			return
		}
		fmt.Printf("request : %v\n", request)
		var response AgentResponse
		var skipResponse bool
		switch request.Op {
		case "cmd":
			response = startCmd(request)
		case "push_file":
			response, skipResponse = pushFile(request, conn)
		case "pull_file":
			response, skipResponse = pullFile(request, conn)
		case "find_log_pos":
			response = findLogPos(request)
		case "find_log":
			response = findLog(request)
		case "find_decider_action":

			//TODO
		}
		if !skipResponse {
			fmt.Printf("response : %v\n", response)
			gob.NewEncoder(conn).Encode(response)
		}
	})
}

func startCmd(request AgentRequest) (response AgentResponse) {
	output, retCode, err := common.Cmd(request.Command)
	if nil != err {
		fmt.Printf("cmd %v got err %v\n", request.Command, err)
		response.ExitCode = 1
		response.Output = err.Error()
		return response
	}
	response.Output = output
	response.ExitCode = retCode
	return response
}

func pullFile(request AgentRequest, conn net.Conn) (response AgentResponse, skipResponse bool) {
	file, err := ioutil.ReadFile(request.FilePath)
	if nil != err {
		response.ExitCode = 1
		response.Output = err.Error()
		return response, false
	}
	fileSize := int64(len(file))
	response.ExitCode = 0
	response.FileSize = fileSize
	gob.NewEncoder(conn).Encode(response)
	var a string
	gob.NewDecoder(conn).Decode(&a)
	conn.Write(file)
	return AgentResponse{}, true

}

func pushFile(request AgentRequest, conn net.Conn) (response AgentResponse, skipResponse bool) {
	var n int
	gob.NewEncoder(conn).Encode("succeed")
	buf := make([]byte, 10000)
	os.Remove(request.FilePath)
	f, err := os.OpenFile(request.FilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if nil != err {
		fmt.Printf("ERROR : %v\n", err)
		return AgentResponse{}, true
	}
	defer f.Close()
	f.Seek(0, 2)
	leftSize := request.FileSize
	for leftSize > 0 {
		n, err = conn.Read(buf)
		if nil != err {
			break
		}
		f.Write(buf[:n])
		leftSize -= int64(n)
	}
	f.Chmod(0755)
	return AgentResponse{}, true
}

func findLog(request AgentRequest) (response AgentResponse) {
	fi, err := os.Open(request.FilePath)
	defer fi.Close()
	if err != nil {
		response.ExitCode = 1
		response.Output = fmt.Sprintf("find log open file %v err : %v", request.FilePath, err)
		return response
	}
	if _, err := fi.Seek(request.LogStartPos, 0); nil == err {
		response.LogCheckpointPos = -1
	} else {
		response.LogCheckpointPos = 0
	}
	regex := regexp.MustCompile(request.LogExpectRegexp)
	r := bufio.NewReader(fi)
	for {
		line, err := r.ReadString('\n')
		if nil != err {
			break
		}
		if regex.MatchString(line) {
			response.ExitCode = 0
			return response
		}
	}
	response.ExitCode = 1
	return response
}

func findLogPos(request AgentRequest) (response AgentResponse) {
	stat, err := os.Stat(request.FilePath)
	if nil != err {
		response.ExitCode = 1
		response.Output = fmt.Sprintf("find log pos of %v err : %v", request.FilePath, err)
		return response
	}
	response.LogCheckpointPos = stat.Size()
	return response
}

func findDeciderAction(request AgentRequest) (response AgentResponse) {
	return response
}
