package main

import (
	"auto-qa/nos-cli/cli"
)

func main() {
	cli.startCmd("127.0.0.1", "5555", "ls")
}
