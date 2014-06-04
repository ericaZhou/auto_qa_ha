package ha

import (
	"fmt"
	"regexp"
	"strings"
)

func (t *Testcase) varMasterSlave(matches []string) error {
	m, s, err := t.getMasterSlave()
	fmt.Errorf("in variables")
	if nil != err {
		return err
	}
	t.variables["master"] = m
	t.variables["slave"] = s
	return nil
}

func (t *Testcase) varNodes(matches []string) error {
	primaryIps := ""
	if output, err := t.getHaStatus(); nil != err {
		return err
	} else {

		hacStatus := strings.Split(output, "Global Status")
		for i := 1; i < len(hacStatus); i++ {
			status := hacStatus[i]
			fmt.Printf("\ni = %v ,status = %v\n", i, status)
			master := ""
			slave := ""
			ip := ""
			for _, ip = range []string{t.variables["a"], t.variables["b"], t.variables["c"], t.variables["d"]} {
				if regexp.MustCompile(fmt.Sprintf("%v.*with SIP", ip)).MatchString(status) {
					if "" != master {
						return fmt.Errorf("Found two master, status was~~~~~: %v", output)
					}
					master = ip
				} else if strings.Contains(status, ip) {
					if "" != slave {
						return fmt.Errorf("Found two slave, status was: %v", output)
					}
					slave = ip
				}

			}
			fmt.Printf("~~~~~~~~~~~~~~~~~~slave=%v", slave)

			fmt.Printf("~~~~~~~~~~~~~~~~~~master=%v,hacStatus[0]=  %v: ,result=%t ", master, hacStatus[0], strings.Contains(hacStatus[0], master))
			if strings.Contains(hacStatus[0], master) {
				primaryIps = primaryIps + " "
				t.variables["pm"] = master
				t.variables["ps"] = slave
			} else {

				t.variables["sm"] = master
				t.variables["ss"] = slave
			}
			if "" == master {
				return fmt.Errorf("Found no master, status was: %v", output)
			}
			if "" == slave {
				return fmt.Errorf("Found no slave, status was: %v", output)
			}
			master, slave = "", ""

		}
		if "" == primaryIps {
			fmt.Errorf("Found no primary, status was: %v", output)
		}
	}

	t.lookVars()
	return nil
}
func (t *Testcase) varSip(matches []string) error {
	t.variables["m_sip"] = t.GetConfigString("remote", "m_sip")
	t.variables["s_sip"] = t.GetConfigString("remote", "s_sip")
	return nil
}

func (t *Testcase) varNic(matches []string) error {
	t.variables["nic"] = t.GetConfigString("remote", "nic")
	return nil
}

func (t *Testcase) setSystemConfigVar(matches []string) error {
	ipA, _ := t.getVar("A")
	ipB, _ := t.getVar("B")
	for _, ip := range []string{ipA, ipB} {
		t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
			if "set" == matches[1] {
				output = input + "\n" + fmt.Sprintf("set system config %v.%v=%v", matches[2], matches[3], matches[4])
			} else {
				output = strings.Replace(input, fmt.Sprintf("set system config %v.%v=%v", matches[2], matches[3], matches[4]), "", -1)
			}
			return output, nil
		})
	}
	return nil
}
func (t *Testcase) vars(matches []string) error {
	varName, ip := matches[1], matches[2]
	t.variables[varName] = ip
	return nil
}

func (t *Testcase) setSenariaVars() {
	sections := t.config.GetSections()
	for _, section := range sections {
		options, _ := t.config.GetOptions(section)
		for _, option := range options {
			t.variables[option], _ = t.config.GetString(section, option)
		}
	}
}
func (t *Testcase) lookVars() {
	for s, i := range t.variables {
		fmt.Print("s:" + s)
		fmt.Println(" i :" + i)
	}
}
