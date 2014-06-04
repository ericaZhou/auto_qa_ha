package main

import (
	"auto-qa/nos-cli/agent"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(50)
	agent.StartAgentServer()
	for {
		time.Sleep(1 * time.Minute)
	}
}
