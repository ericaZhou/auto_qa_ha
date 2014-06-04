package main

import (
	"auto-qa/dha"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

func main() {
	flag.Parse()
	os.RemoveAll("result")
	os.Mkdir("result", 0700)
	scenario := flag.Arg(0)
	haZip := flag.Arg(1)
	cases := findCases(flag.Arg(2))
fmt.Println(flag.Arg(2))
fmt.Println(cases)
	report := fmt.Sprintf("<testsuite tests=\"%v\">", len(cases))
	for _, c := range cases {
		if pass, seconds := runTestcase(scenario, haZip, fmt.Sprintf("../testcases/%v.tc", c)); pass {
			log, _ := ioutil.ReadFile(fmt.Sprintf("./result/%v/log", c))
			var buf bytes.Buffer
			xml.EscapeText(&buf, log)
			report = report + fmt.Sprintf("\n<testcase classname=\"%v\" name=\"%v\" time=\"%v\">\n<system-out>%v</system-out>\n</testcase>", "HA", c, seconds, buf.String())
		} else {
			log, _ := ioutil.ReadFile(fmt.Sprintf("./result/%v/log", c))
			var buf bytes.Buffer
			xml.EscapeText(&buf, log)
			report = report + fmt.Sprintf("\n<testcase classname=\"%v\" name=\"%v\" time=\"%v\">\n<failure type=\"CaseFail\"/>\n<system-out>%v</system-out>\n</testcase>", "default", c, seconds, buf.String())
		}
	}
	report = report + "\n</testsuite>"
	ioutil.WriteFile("./result/test_report.xml", []byte(report), 0644)
}

func runTestcase(scenario, haZip, path string) (pass bool, seconds int) {
	fmt.Printf("Running %v: ", path)
	tc := ha.NewTestcase(scenario, haZip, path)
	startTime := time.Now()
	err := tc.Run()
	spentTime := int(time.Since(startTime).Seconds())
	if nil != err {
		fmt.Println("fail")
		fmt.Printf("\t%v, %vs\n", err, spentTime)
		return false, spentTime
	} else {
		fmt.Printf("pass, %vs\n", spentTime)
		return true, spentTime
	}
	return true, 0
}
