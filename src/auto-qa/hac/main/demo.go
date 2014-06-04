package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	action := "run decider"
	reg := "^run decider"

	r := regexp.MustCompile("(?s)" + reg)

	matches := r.FindStringSubmatch(action)

	fmt.Println(matches)

	hacStatus := "HAC primary: 192.168.56.101, 192.168.56.102\\n\\n\\nGlobal Status -- Running --- 192.168.56.223\\n|\\n+-- 192.168.56.103(eth1) with SIP\\n+-- HA Service: Running\\n+-- MySQL Service: Running\\n+-- Role: Master\\n|\\n+-- 192.168.56.104(eth1)\\n+-- HA Service: Running\\n+-- MySQL Service: Running\\n+-- MySQL Replication: Running\\n+-- Role: Slave\\n\\nGlobal Status -- Running --- 192.168.56.222\\n|\\n+-- 192.168.56.101(eth1) with SIP\\n+-- HA Service: Running\\n+-- MySQL Service: Running\\n+-- Role: Master\\n|\\n+-- 192.168.56.102(eth1)\\n+-- HA Service: Running\\n+-- MySQL Service: Running\\n+-- MySQL Replication: Running\\n+-- Role: Slave"
	fmt.Println(hacStatus)
	st := strings.Split(hacStatus, "Global Status")

	fmt.Println(strings.Contains(st[0], "192.168.56.101"))
	for _, s := range st {

		fmt.Println(s)
		fmt.Println()

	}
}

func findCases(numberStr string) []string {
	numbers := make([]int, 0)
	if "" == numberStr {
		for i := 0; i < 1000; i++ {
			numbers = append(numbers, i)
		}
	} else if regexp.MustCompile("\\d+-\\d+").MatchString(numberStr) {
		match := regexp.MustCompile("(\\d+)-(\\d+)").FindStringSubmatch(numberStr)
		low, _ := strconv.Atoi(match[1])
		high, _ := strconv.Atoi(match[2])
		for i := low; i <= high; i++ {
			numbers = append(numbers, i)
		}
	} else {
		for _, i := range strings.Split(numberStr, ",") {
			a, _ := strconv.Atoi(i)
			numbers = append(numbers, a)
		}
	}
	ret := make([]string, 0)
	filepath.Walk("../testcases", func(path string, f os.FileInfo, err error) error {
		if ".tc" == filepath.Ext(path) {
			a, _ := strconv.Atoi(filepath.Base(path)[0:3])
			for _, i := range numbers {
				if a == i {
					ret = append(ret, strings.TrimSuffix(filepath.Base(path), ".tc"))
					break
				}
			}
		}
		return nil
	})
	return ret
}

func parseTime(a string) int {
	var timeout int
	if strings.HasSuffix(a, "min") {
		timeout, _ = strconv.Atoi(strings.TrimSuffix(a, "min"))
		timeout = timeout * 60
	} else if strings.HasSuffix(a, "s") {
		timeout, _ = strconv.Atoi(strings.TrimSuffix(a, "s"))
	}
	return timeout
}
