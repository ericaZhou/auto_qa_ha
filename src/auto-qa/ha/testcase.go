package ha

import (
	"auto-qa/nos-cli/cli"
	"fmt"
	"goconf/conf"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type Testcase struct {
	scenario        string
	filePath        string
	variables       map[string]string
	logCheckpoint   map[string]int64
	config          *conf.ConfigFile
	haStatus        string
	haZip           string
	logDir          string
	keepRunQuitChan chan bool
	bgDoneChan      chan bool
}

func NewTestcase(scenario, haZip, filePath string) *Testcase {
	ret := Testcase{}
	ret.scenario = scenario
	ret.filePath = filePath
	ret.haZip = haZip
	ret.variables = make(map[string]string)
	ret.logCheckpoint = make(map[string]int64)
	ret.keepRunQuitChan = make(chan bool, 10)
	return &ret
}

func (t *Testcase) Run() (err error) {
	defer func() {
		t.keepRunQuitChan <- true
	}()

	//set logger
	t.logDir = "result/" + strings.TrimSuffix(path.Base(t.filePath), ".tc")
	os.Mkdir(t.logDir, 0700)
	f, err := os.OpenFile(t.logDir+"/log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	//init config
	t.config, err = conf.ReadConfigFile(t.scenario)
	if nil != err {
		return err
	}
	cli.InitRemoteOs(t.GetConfigString("os", "os"))

	//var A,B
	t.doAction(fmt.Sprintf("var A = %v", t.GetConfigString("remote", "a")))
	t.doAction(fmt.Sprintf("var B = %v", t.GetConfigString("remote", "b")))

	bytes, err := ioutil.ReadFile(t.filePath)
	if nil != err {
		return err
	}
	lastLine := ""
	isMultiLine := false
	stringParam := ""
	inRollback := false

	for _, line := range strings.Split(string(bytes), "\n") {
		if line == "rollback:" {
			inRollback = true
			log.Println("ROLLBACK:")
			continue
		}
		if nil != err && !inRollback {
			continue
		}
		if strings.HasSuffix(line, "`") {
			if !isMultiLine {
				isMultiLine = true
				lastLine = strings.TrimSuffix(line, "`")
				stringParam = ""
				continue
			} else {
				line = fmt.Sprintf("%v\"%v\"", lastLine, strings.TrimPrefix(stringParam, "\n"))
				isMultiLine = false
			}
		}
		if isMultiLine {
			stringParam = fmt.Sprintf("%v\n%v", stringParam, strings.TrimSpace(line))
			continue
		}
		if inRollback {
			t.doAction(line)
			continue
		}
		if err = t.doAction(line); nil != err {
			err = fmt.Errorf("case %v failed @\"%v\"", t.filePath, line)
		}
	}
	if nil != err {
		t.onFailure()
	}
	return err
}

func (t *Testcase) isWindows() bool {
	return "windows" == t.GetConfigString("os", "os")
}

func (t *Testcase) GetConfigString(section, option string) string {
	a, err := t.config.GetString(section, option)
	if nil != err {
		panic(err)
	}
	return a
}
func (t *Testcase) GetConfigInt(section, option string) int {
	a, err := t.config.GetInt(section, option)
	if nil != err {
		panic(err)
	}
	return a
}
