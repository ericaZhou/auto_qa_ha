package ha

import (
	"auto-qa/common"
	"auto-qa/nos-cli/cli"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (t *Testcase) doAction(line string) error {
	for desc, action := range t.getActions() {
		r := regexp.MustCompile("(?s)" + desc)
		if !r.MatchString(line) {
			continue
		}
		matches := r.FindStringSubmatch(line)
		log.Printf("%v> %v\n", time.Now(), line)
		if err := action(matches); nil != err {
			log.Printf("err>> %v\n", err)
			return err
		}
		return nil
	}
	return nil
}

func (t *Testcase) getActions() map[string]func([]string) error {
	ret := make(map[string]func([]string) error)
	ret["^var (\\w+) = ([^ ]+)"] = t.vars
	ret["^var sip"] = t.varSip
	ret["^var nic"] = t.varNic
	ret["^var master slave"] = t.varMasterSlave
	ret["^install HA on \\{(\\w+)\\}"] = t.installHa
	ret["^extract HA installation on \\{(\\w+)\\}"] = t.extractHaInstallation
	ret["^run HA install script on \\{(\\w+)\\}"] = t.runHaInstallScript
	ret["^assert \\{(\\w+)\\} log \"(.+)\", timeout=(.+)"] = t.assertLog
	ret["^assert \\{(\\w+)\\} log \"(.+)\"$"] = t.assertLog
	ret["^assert \\{(\\w+)\\} no log \"(.+)\"$"] = t.assertNoLog
	ret["^wait \\{(\\w+)\\} log \"(.+)\", timeout=(.+)"] = t.assertLog
	ret["^wait \\{(\\w+)\\} log \"(.+)\"$"] = t.assertLog
	ret["^clean env on \\{(\\w+)\\}"] = t.cleanEnv
	ret["^get status"] = t.getStatus
	ret["^assert status has \"(.+)\""] = t.assertStatusHas
	ret["^assert status all ok"] = t.assertStatusAllOk
	ret["^print status"] = t.printStatus
	ret["^prepare clean env"] = t.prepareCleanEnv
	ret["^(start|stop) \\{(.+)\\} ha service"] = t.startStopHaService
	ret["^kill \\{(.+)\\} ha service"] = t.killHaService
	ret["^(start|stop) \\{(.+)\\} mysql service"] = t.startStopMysqlService
	ret["^stop \\{(.+)\\} http service"] = t.stopHttpService
	ret["^kill \\{(.+)\\} mysqld"] = t.killMysqld
	ret["^set \\{(.+)\\} log checkpoint"] = t.setLogCheckpoint
	ret["^run \\{(.+)\\} mysql \"(.+)\""] = t.runMysql
	ret["^assert \\{(.+)\\} mysql \"(.+)\" output is \"(.+)\"$"] = t.assertMysqlOutput
	ret["^assert \\{(.+)\\} is (master|slave)"] = t.assertMasterSlave
	ret["^wait random time < (\\d+)s"] = t.waitRandomTime
	ret["^clean \\{(.+)\\} unstable count"] = t.cleanUnstableCount
	ret["^del \\{(.+)\\} mysql data folder"] = t.delMysqlDataFolder
	ret["^del \\{(.+)\\} mysql folder"] = t.delMysqlFolder
	ret["^disable \\{(.+)\\} network for (.+)"] = t.disableNic
	ret["^wait \\{(.+)\\} network resume, timeout=(.+)"] = t.waitNicResume
	ret["^sysbench prepare \\{(.+)\\} (\\d+)w data"] = t.sysbenchPrepareData
	ret["^wait \\{(.+)\\} mysql data catch up with \\{(.+)\\}, timeout=(.+)"] = t.waitMysqlDataCatchUp
	ret["^keep run \\{(.+)\\} mysql \"(.+)\""] = t.keepRunMysql
	ret["^stop keep run"] = t.stopKeepRun
	ret["^assert \\{(.+)\\} sysbench has (\\d+)w data"] = t.assertHasSysbenchData
	ret["^drop \\{(.+)\\} sysbench data"] = t.dropSysbenchData
	ret["^keep run \\{(.+)\\} kill mysqld"] = t.keepRunKillMysqld
	ret["^sleep (.+)"] = t.sleep
	ret["^assert connect \\{(.+)\\} mysql should succeed"] = t.assertConnectMysql
	ret["^bg (.+)"] = t.bg
	ret["^wait bg"] = t.waitBg
	ret["^kill \\{(.+)\\} mysql binlog dump process"] = t.killMysqlBinlogDumpProcess
	ret["^set promote_limit_seconds = (\\d+)"] = t.setPromoteLimitSeconds
	ret["^(unset|set) \\{(.+)\\} recovery fail"] = t.setRecoveryFail
	ret["^del \\{(.+)\\} node_lock"] = t.delNodeLock
	ret["^(unset|set) system config (.+) = (.+)"] = t.setSystemConfigVar
	ret["^enable app scripts"] = t.enableAppScripts
	ret["^set \\{(.+)\\} app availability (false|true)"] = t.setAppAvailability
	ret["^disable app scripts"] = t.disableAppScripts
	ret["^add \\{(.+)\\} init-db-script \"(.+)\""] = t.addInitDbScript
	ret["^del \\{(.+)\\} mysql backup"] = t.delMysqlBackup
	ret["^(break|recover) \\{(.+)\\} tcp"] = t.breakRecoverTcp
	ret["^assert \\{(.+)\\} app before (master|slave) ready script run"] = t.assertAppBeforeMsReadyScriptRun
	return ret
}

func (t *Testcase) assertAppBeforeMsReadyScriptRun(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, retCode, err := cli.Cmdf(ip, "ls app_%v_ready_script_run", matches[2]); nil != err || 0 == retCode {
		return fmt.Errorf("assert app_%v_ready_script_run but failed, err=%v, ls retCode=%v", err, retCode)
	}
	return nil
}

func (t *Testcase) breakRecoverTcp(matches []string) error {
	ip, err := t.getVar(matches[2])
	if nil != err {
		return err
	}
	t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
		if matches[1] == "break" {
			output = input + "\n" + "break tcp"
		} else {
			output = strings.Replace(input, "break tcp", "", -1)
		}
		return output, nil
	})
	return nil
}

func (t *Testcase) delMysqlBackup(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.RmDir(ip, t.GetConfigString("remote_path", "ha_dir")+"/mysql-backup"); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) addInitDbScript(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	tempFile := path.Join(os.TempDir(), "tmp.sql")
	defer os.Remove(tempFile)
	ioutil.WriteFile(tempFile, []byte(matches[2]), 0755)
	if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/init_db_script.sql"); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) stopHttpService(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	cli.Cmdf(ip, "ps aux |grep [h]ttp_server |awk '{print $2}' |xargs kill -9")
	return nil
}

func (t *Testcase) setAppAvailability(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	tempFile := path.Join(os.TempDir(), "test.sh")
	defer os.Remove(tempFile)
	ioutil.WriteFile(tempFile, []byte(fmt.Sprintf("echo {\\\"availability\\\":\\\"%v\\\"}", matches[2])), 0755)
	if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_availability_script.sh"); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) enableAppScripts(matches []string) error {
	ipA, _ := t.getVar("A")
	ipB, _ := t.getVar("B")
	for _, ip := range []string{ipA, ipB} {
		t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
			output = input
			output = fmt.Sprintf("%v\nset node config app_start_script=\"%v/app_start_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			output = fmt.Sprintf("%v\nset node config app_stop_script=\"%v/app_stop_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			output = fmt.Sprintf("%v\nset node config app_availability_script=\"%v/app_availability_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			output = fmt.Sprintf("%v\nset node config app_status_check_script=\"%v/app_status_check_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			output = fmt.Sprintf("%v\nset node config app_before_master_ready_script=\"%v/app_before_master_ready_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			output = fmt.Sprintf("%v\nset node config app_before_slave_ready_script=\"%v/app_before_slave_ready_script.sh\"", output, t.GetConfigString("remote_path", "ha_dir"))
			return output, nil
		})
		tempFile := path.Join(os.TempDir(), "test.sh")
		defer os.Remove(tempFile)
		ioutil.WriteFile(tempFile, []byte("echo {\\\"is_started\\\":\\\"true\\\"}"), 0755)
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_start_script.sh"); nil != err {
			return err
		}
		ioutil.WriteFile(tempFile, []byte("echo {\\\"is_stopped\\\":\\\"true\\\"}"), 0755)
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_stop_script.sh"); nil != err {
			return err
		}
		ioutil.WriteFile(tempFile, []byte("echo {\\\"availability\\\":\\\"true\\\"}"), 0755)
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_availability_script.sh"); nil != err {
			return err
		}
		ioutil.WriteFile(tempFile, []byte("echo {\\\"status\\\":\\\"started\\\"}"), 0755)
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_status_check_script.sh"); nil != err {
			return err
		}
		ioutil.WriteFile(tempFile, []byte("echo 1 > /tmp/app_master_ready_script_run; echo {\\\"is_started\\\":\\\"true\\\"}"), 0755)
		cli.Rm(ip, "/tmp/app_master_ready_script_run")
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_before_master_ready_script.sh"); nil != err {
			return err
		}
		ioutil.WriteFile(tempFile, []byte("echo 1 > /tmp/app_slave_ready_script_run; echo {\\\"is_started\\\":\\\"true\\\"}"), 0755)
		cli.Rm(ip, "/tmp/app_slave_ready_script_run")
		if err := cli.PushFile(ip, tempFile, t.GetConfigString("remote_path", "ha_dir")+"/app_before_slave_ready_script.sh"); nil != err {
			return err
		}
	}
	if err := t.restartAndWaitAllNodes(); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) restartAndWaitAllNodes() error {
	if err := t.doAction("set {A} log checkpoint"); nil != err {
		return err
	}
	if err := t.doAction("set {B} log checkpoint"); nil != err {
		return err
	}
	if err := t.doAction("stop {A} ha service"); nil != err {
		return err
	}
	if err := t.doAction("stop {B} ha service"); nil != err {
		return err
	}
	if err := t.doAction("stop {A} http service"); nil != err {
		return err
	}
	if err := t.doAction("stop {B} http service"); nil != err {
		return err
	}
	if err := t.doAction("start {A} ha service"); nil != err {
		return err
	}
	if err := t.doAction("start {B} ha service"); nil != err {
		return err
	}
	if err := t.doAction("wait {A} log \"Node is running\", timeout=3min"); nil != err {
		return err
	}
	if err := t.doAction("wait {B} log \"Node is running\", timeout=3min"); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) disableAppScripts(matches []string) error {
	ipA, _ := t.getVar("A")
	ipB, _ := t.getVar("B")
	for _, ip := range []string{ipA, ipB} {
		t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
			//FIXME clean set node config instead of clean all
			return "", nil
		})
	}
	if err := t.restartAndWaitAllNodes(); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) setSystemConfigVar(matches []string) error {
	ipA, _ := t.getVar("A")
	ipB, _ := t.getVar("B")
	for _, ip := range []string{ipA, ipB} {
		t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
			if "set" == matches[1] {
				output = input + "\n" + fmt.Sprintf("set system config %v=%v", matches[2], matches[3])
			} else {
				output = strings.Replace(input, fmt.Sprintf("set system config %v=%v", matches[2], matches[3]), "", -1)
			}
			return output, nil
		})
	}
	return nil
}

func (t *Testcase) delNodeLock(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	var path string
	if t.isWindows() {
		path = t.GetConfigString("remote_path", "ha_dir") + "\\node_lock"
	} else {
		path = t.GetConfigString("remote_path", "ha_dir") + "/node_lock"
	}
	if _, _, err := cli.Rm(ip, path); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) setRecoveryFail(matches []string) error {
	action := matches[1]
	ip, err := t.getVar(matches[2])
	if nil != err {
		return err
	}
	t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
		if "set" == action {
			output = fmt.Sprintf("%v\nrecovery fail", input)
			return output, nil
		} else {
			output = strings.Replace(input, "\nrecovery fail", "", -1)
			return output, nil
		}
	})
	return nil
}

func (t *Testcase) setPromoteLimitSeconds(matches []string) error {
	ipA, _ := t.getVar("A")
	ipB, _ := t.getVar("B")
	for _, ip := range []string{ipA, ipB} {
		t.updateAutoQaDebug(ip, func(input string) (output string, err error) {
			output = fmt.Sprintf("%v\npromote_limit_seconds=%v", input, matches[1])
			return output, nil
		})
	}
	return nil
}

func (t *Testcase) killMysqlBinlogDumpProcess(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	output, retCode, err := cli.Cmdf(ip, "%v -N -B -e \"%v\"",
		t.GetConfigString("remote_command", "mysql"),
		"show processlist")
	if nil != err || 0 != retCode {
		return fmt.Errorf("show processlist failed, err=%v, retCode=%v", err, retCode)
	}
	match := regexp.MustCompile("^(\\d+).*Binlog Dump").FindStringSubmatch(output)
	if nil == match {
		return nil
	}
	cli.Cmdf(ip, "%v -e \"%v\"",
		t.GetConfigString("remote_command", "mysql"),
		"kill "+match[1])
	return nil
}

func (t *Testcase) assertMysqlOutput(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	output, retCode, err := cli.Cmdf(ip, "%v -N -B -e \"%v\"",
		t.GetConfigString("remote_command", "mysql"),
		matches[2])
	if nil != err || 0 != retCode {
		return fmt.Errorf("assert mysql output failed, err=%v, retCode=%v, output=%v", err, retCode, output)
	}
	output = strings.TrimSpace(output)
	if matches[3] != output {
		return fmt.Errorf("assert mysql output expect %v, but got %v", matches[3], output)
	}
	return nil
}

func (t *Testcase) waitBg(matches []string) error {
	if nil != t.bgDoneChan {
		<-t.bgDoneChan
	}
	t.bgDoneChan = nil
	return nil
}

func (t *Testcase) bg(matches []string) error {
	if nil != t.bgDoneChan {
		return fmt.Errorf("Already has bf task")
	}
	t.bgDoneChan = make(chan bool, 1)
	go func(a string) {
		t.doAction(a)
		t.bgDoneChan <- true
	}(matches[1])
	return nil
}

func (t *Testcase) killHaService(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if t.isWindows() {
		//TODO windows
	} else {
		if _, _, err := cli.Cmdf(ip, "ps aux | grep [c]om.actionsky.ha.Bootstrap | awk '{print $2}' | xargs kill -9"); nil != err {
			return err
		}
	}
	return nil
}

func (t *Testcase) assertConnectMysql(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	_, retCode, err := cli.Cmdf(ip, "%v -e \"select 1\"", t.GetConfigString("remote_command", "mysql"))
	if nil != err || 0 != retCode {
		return fmt.Errorf("assert connect mysql failed, err=%v, retCode=%v", err, retCode)
	}
	return nil
}

func (t *Testcase) sleep(matches []string) error {
	a := t.parseTime(matches[1])
	time.Sleep(time.Duration(a) * time.Second)
	return nil
}

func (t *Testcase) keepRunKillMysqld(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	go func(ip string) {
		for {
			cli.Cmdf(ip, "killall mysqld")
			select {
			case <-t.keepRunQuitChan:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(ip)
	return nil
}

func (t *Testcase) dropSysbenchData(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	cli.Cmdf(ip, fmt.Sprintf("%v -e \"drop table sysbench.sbtest\"", t.GetConfigString("remote_command", "mysql")))
	return nil
}

func (t *Testcase) assertHasSysbenchData(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	expectCount := matches[2] + "0000"
	if output, retCode, err := cli.Cmdf(ip, fmt.Sprintf("%v -e \"select count(1) from sysbench.sbtest\\G\"", t.GetConfigString("remote_command", "mysql"))); nil != err || 0 != retCode {
		return fmt.Errorf("select count(1) got error=%v, retCode=%v", err, retCode)
	} else {
		if strings.Contains(output, expectCount) {
			return nil
		}
	}
	return fmt.Errorf("assertHasSysbenchData failed")
}

func (t *Testcase) stopKeepRun(matches []string) error {
	t.keepRunQuitChan <- true
	return nil
}

func (t *Testcase) keepRunMysql(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	mysql := matches[2]
	go func(ip, mysql string) {
		for {
			cli.Cmdf(ip, fmt.Sprintf("%v -e \"%v\"", t.GetConfigString("remote_command", "mysql"), mysql))
			select {
			case <-t.keepRunQuitChan:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(ip, mysql)
	return nil
}

func (t *Testcase) waitMysqlDataCatchUp(matches []string) error {
	slaveIp, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	masterIp, err := t.getVar(matches[2])
	if nil != err {
		return err
	}
	timeout := t.parseTime(matches[3])
	timeWeight := t.GetConfigInt("remote", "time_weight")
	for i := 0; i <= timeout*timeWeight; i++ {
		masterPos := ""
		if output, retCode, err := cli.Cmdf(masterIp, fmt.Sprintf("%v -e \"show master status\\G\"", t.GetConfigString("remote_command", "mysql"))); nil != err || 0 != retCode {
			return fmt.Errorf("show master status got error=%v, retCode=%v", err, retCode)
		} else {
			match := regexp.MustCompile("Position: (\\d+)").FindStringSubmatch(output)
			if len(match) != 2 {
				return fmt.Errorf("show master status has wrong output : %v", output)
			}
			masterPos = match[1]
		}
		if output, retCode, err := cli.Cmdf(slaveIp, fmt.Sprintf("%v -e \"show slave status\\G\"", t.GetConfigString("remote_command", "mysql"))); nil == err && 0 == retCode {
			match := regexp.MustCompile("Exec_Master_Log_Pos: (\\d+)").FindStringSubmatch(output)
			if len(match) != 2 {
				continue
			}
			slavePos := match[1]
			if masterPos == slavePos {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("wait mysql data catch up failed")
}

func (t *Testcase) sysbenchPrepareData(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	cli.Cmdf(ip, "%v -e \"create database if not exists sysbench\"", t.GetConfigString("remote_command", "mysql"))
	cli.Cmdf(ip, "%v -e \"drop table if exists sysbench.sbtest\"", t.GetConfigString("remote_command", "mysql"))
	volume, _ := strconv.Atoi(matches[2])
	for i := 0; i < 3; i++ {
		cmd := fmt.Sprintf("%v --num-threads=200 --test=oltp --mysql-host=%v --mysql-port=%v --mysql-user=root --mysql-password=root --mysql-db=sysbench --mysql-table-engine=innodb --oltp-read-only=on --max-requests=0 --max-time=7200 --oltp-test-mode=complex --oltp-table-size=%v0000 prepare",
			t.GetConfigString("local_path", "sysbench"), ip, t.GetConfigString("remote", "mysql_port"), volume)
		output, retCode, err := common.Cmd(cmd)
		if nil == err && 0 == retCode {
			return nil
		}
		if 2 == i {
			return fmt.Errorf("sysbench failed, err=%v, retCode=%v, output=%v", err, retCode, output)
		}
	}
	return nil
}

func (t *Testcase) waitNicResume(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	timeout := t.parseTime(matches[2])
	timeWeight := t.GetConfigInt("remote", "time_weight")
	for i := 0; i <= timeout*timeWeight; i++ {
		if _, exitCode, err := cli.Cmdf(ip, "ifconfig"); nil != err || 0 != exitCode {
			time.Sleep(1 * time.Second)
			continue
		} else {
			return nil
		}
	}
	return fmt.Errorf("wait network resume timeout")
}

func (t *Testcase) disableNic(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	timeout := t.parseTime(matches[2])

	cli.Cmdf(ip, fmt.Sprintf("/etc/init.d/network stop && sleep %v && /etc/init.d/network start", timeout)) //ignore error
	return nil
}

func (t *Testcase) updateAutoQaDebug(ip string, fn func(input string) (output string, err error)) error {
	autoQaDebugPath := path.Join(t.GetConfigString("remote_path", "ha_dir"), "auto_qa.debug")
	tempFile := path.Join(os.TempDir(), "auto_qa.debug")
	defer os.Remove(tempFile)

	if err := cli.PullFile(ip, autoQaDebugPath, tempFile); nil != err {
		return err
	}

	a, err := ioutil.ReadFile(tempFile)
	if nil != err {
		return err
	}
	output, err := fn(string(a))
	if nil != err {
		return err
	}

	err = ioutil.WriteFile(tempFile, []byte(output), 0644)
	if nil != err {
		return err
	}
	if err := cli.PushFile(ip, tempFile, autoQaDebugPath); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) delMysqlFolder(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.RmDir(ip, path.Join(t.GetConfigString("remote_path", "mysql_install"))); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) delMysqlDataFolder(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.RmDir(ip, path.Join(t.GetConfigString("remote_path", "mysql_install"), "data")); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) killMysqld(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.Cmdf(ip, "killall mysqld"); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) cleanUnstableCount(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	err = t.updateAutoQaDebug(ip, func(input string) (string, error) {
		output := input + "\nclean unstable count"
		return output, nil
	})
	if nil != err {
		return err
	}
	err = t.doAction(fmt.Sprintf("set {%v} log checkpoint", matches[1]))
	if nil != err {
		return err
	}
	err = t.doAction(fmt.Sprintf("wait {%v} log \"auto_qa.debug : clean unstable count done\", timeout=1min", matches[1]))
	if nil != err {
		return err
	}
	err = t.updateAutoQaDebug(ip, func(input string) (string, error) {
		output := strings.Replace(input, "\nclean unstable count", "", -1)
		return output, nil
	})
	if nil != err {
		return err
	}
	return nil
}

func (t *Testcase) waitRandomTime(matches []string) error {
	seconds, _ := strconv.Atoi(matches[1])
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func (t *Testcase) runMysql(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.Cmdf(ip, fmt.Sprintf("%v -e \"%v\"", t.GetConfigString("remote_command", "mysql"), matches[2])); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) isHaStatusAllOk(status string) bool {
	return (2 == len(regexp.MustCompile("HA Service: Running").FindAllString(status, -1))) &&
		(2 == len(regexp.MustCompile("MySQL Service: Running").FindAllString(status, -1))) &&
		(1 == len(regexp.MustCompile("MySQL Replication: Running").FindAllString(status, -1)))
}

func (t *Testcase) assertMasterSlave(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, exitCode, err := cli.Cmdf(ip, "ps aux | grep [m]ysqld_safe"); nil == err && 0 == exitCode {
		return fmt.Errorf("Found mysqld_safe on %v, which is illegal", ip)
	}
	m, s, err := t.getMasterSlave()
	if nil != err {
		return err
	}
	hasSip, err := t.hasIp(ip, t.GetConfigString("remote", "sip"))
	if nil != err {
		return err
	}
	showSlaveStatus := ""
	if output, exitCode, err := cli.Cmdf(ip, fmt.Sprintf("%v -e \"%v\"", t.GetConfigString("remote_command", "mysql"), "show slave status\\G")); nil != err || 0 != exitCode {
		return fmt.Errorf("show slave status got error : %v, %v, output=%v", err, exitCode, output)
	} else {
		showSlaveStatus = output
	}

	if "master" == matches[2] {
		if m != ip {
			return fmt.Errorf("assert master failed, expect %v, but got %v", ip, m)
		}
		if !hasSip {
			return fmt.Errorf("assert master has sip, but faild")
		}
		if "" != showSlaveStatus {
			return fmt.Errorf("assert master has no slave status, but failed. got %v", showSlaveStatus)
		}
	} else if "slave" == matches[2] {
		if s != ip {
			return fmt.Errorf("assert slave failed, expect %v, but got %v", ip, s)
		}
		if hasSip {
			return fmt.Errorf("assert slave has no sip, but faild")
		}
		if "" == showSlaveStatus {
			return fmt.Errorf("assert slave has slave status, but failed.")
		}
		if !strings.Contains(showSlaveStatus, "Slave_IO_Running: Yes") {
			return fmt.Errorf("assert slave IO thread is running, but failed. got %v", showSlaveStatus)
		}
		if !strings.Contains(showSlaveStatus, "Slave_SQL_Running: Yes") {
			return fmt.Errorf("assert slave SQL thread is running, but failed. got %v", showSlaveStatus)
		}
		if output, exitCode, err := cli.Cmdf(ip, fmt.Sprintf("%v -e \"%v\"", t.GetConfigString("remote_command", "mysql"), "show variables like '%%read_only%%'\\G")); nil != err || 0 != exitCode {
			return fmt.Errorf("show variables like '%%read_only%%'\\G got error : %v, %v, output=%v", err, exitCode, output)
		} else {
			if !strings.Contains(output, "Value: ON") {
				return fmt.Errorf("assert slave read_only, but failed. got : %v", output)
			}
		}
	}
	return nil
}

func (t *Testcase) hasIp(target, ip string) (bool, error) {
	if t.isWindows() {
		//TODO windows
	} else {
		if _, exitCode, err := cli.Cmdf(target, fmt.Sprintf("ip addr | grep %v/", ip)); nil != err {
			return false, err
		} else {
			return 0 == exitCode, nil
		}
	}
	return false, nil
}

func (t *Testcase) setLogCheckpoint(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	pos, err := cli.FindLogPos(ip, t.GetConfigString("remote_path", "trace_log"))
	if nil != err {
		return err
	}
	t.logCheckpoint[ip] = pos
	return nil
}

func (t *Testcase) startStopMysqlService(matches []string) error {
	op := matches[1]
	ip, err := t.getVar(matches[2])
	if nil != err {
		return err
	}
	if t.isWindows() {
		if "stop" == op {
			if _, _, err := cli.Cmdf(ip, "sc stop ACTIONSKY-HA-Mysql-"+t.GetConfigString("remote", "mysql_port")); nil != err {
				return err
			}
			if _, _, err := cli.Cmdf(ip, "taskkill /im mysqld.exe /f"); nil != err {
				return err
			}
		} else {
			if _, _, err := cli.Cmdf(ip, "sc start ACTIONSKY-HA-Mysql-"+t.GetConfigString("remote", "mysql_port")); nil != err {
				return err
			}
		}

	} else {
		if "stop" == op {
			if _, _, err := cli.Cmdf(ip, "service ACTIONSKY-HA-Mysql-"+t.GetConfigString("remote", "mysql_port")+" stop"); nil != err {
				return err
			}
		} else {
			if _, _, err := cli.Cmdf(ip, "service ACTIONSKY-HA-Mysql-"+t.GetConfigString("remote", "mysql_port")+" start"); nil != err {
				return err
			}
		}
	}
	return nil
}

func (t *Testcase) startStopHaService(matches []string) error {
	op := matches[1]
	ip, err := t.getVar(matches[2])
	if nil != err {
		return err
	}
	cmd := ""
	if t.isWindows() {
		if "stop" == op {
			cmd = "sc stop ACTIONSKY-HA"
		} else {
			cmd = "sc start ACTIONSKY-HA"
		}
	} else {
		if "stop" == op {
			cmd = "service haserver stop"
		} else {
			cmd = "service haserver start"
		}
	}
	_, _, err = cli.Cmdf(ip, cmd)
	return err
}

func (t *Testcase) getHaStatus() (status string, err error) {
	output, exitCode, err := cli.Cmdf(
		t.GetConfigString("remote", "a"), t.GetConfigString("remote_path", "status_script"))
	if nil != err {
		return "", fmt.Errorf("get ha status got err : %v", err)
	}
	if 0 != exitCode {
		return "", fmt.Errorf("get ha status got !0 exit code : %v", exitCode)
	}
	return output, nil
}

func (t *Testcase) prepareCleanEnv(matches []string) error {
	if output, err := t.getHaStatus(); nil == err {
		if t.isHaStatusAllOk(output) {
			a, _ := t.getVar("A")
			b, _ := t.getVar("B")
			t.initAutoHaDebug(a)
			t.initAutoHaDebug(b)
			if _, _, err := cli.Rm(a, t.GetConfigString("remote_path", "ha_dir")+"/node_lock"); nil != err {
				return err
			}
			if _, _, err := cli.Rm(b, t.GetConfigString("remote_path", "ha_dir")+"/node_lock"); nil != err {
				return err
			}
			return nil
		}
	}
	if err := t.doAction("clean env on {A}"); nil != err {
		return err
	}
	if err := t.doAction("clean env on {B}"); nil != err {
		return err
	}
	if err := t.doAction("install HA on {A}"); nil != err {
		return err
	}
	if err := t.doAction("install HA on {B}"); nil != err {
		return err
	}
	if err := t.doAction("assert {A} log \"Node is running\", timeout=3min"); nil != err {
		return err
	}
	if err := t.doAction("assert {B} log \"Node is running\", timeout=3min"); nil != err {
		return err
	}

	if output, err := t.getHaStatus(); nil == err {
		if t.isHaStatusAllOk(output) {
			log.Printf("status : %v\n", output)
			return nil
		}
	}
	return fmt.Errorf("prepare clean env failed")
}

func (t *Testcase) printStatus(matches []string) error {
	if output, err := t.getHaStatus(); nil != err {
		return err
	} else {
		log.Printf("\n--- status --- \n%v\n ---------- \n", output)
	}
	return nil
}

func (t *Testcase) getMasterSlave() (master string, slave string, err error) {
	if output, err := t.getHaStatus(); nil != err {
		return "", "", err
	} else {
		master := ""
		slave := ""
		for _, ip := range []string{t.GetConfigString("remote", "a"), t.GetConfigString("remote", "b")} {
			if regexp.MustCompile(fmt.Sprintf("%v.*with SIP", ip)).MatchString(output) {
				if "" != master {
					return "", "", fmt.Errorf("Found two master")
				}
				master = ip
			} else {
				if "" != slave {
					return "", "", fmt.Errorf("Found two slave")
				}
				slave = ip
			}
		}
		if "" == master {
			return "", "", fmt.Errorf("Found no master")
		}
		if "" == slave {
			return "", "", fmt.Errorf("Found no slave")
		}
		return master, slave, nil
	}
	return "", "", nil
}

func (t *Testcase) varMasterSlave(matches []string) error {
	m, s, err := t.getMasterSlave()
	if nil != err {
		return err
	}
	t.variables["master"] = m
	t.variables["slave"] = s
	return nil
}

func (t *Testcase) varSip(matches []string) error {
	t.variables["sip"] = t.GetConfigString("remote", "sip")
	return nil
}

func (t *Testcase) varNic(matches []string) error {
	t.variables["nic"] = t.GetConfigString("remote", "nic")
	return nil
}

func (t *Testcase) parseVars(a string) string {
	for key, val := range t.variables {
		a = strings.Replace(a, "{"+key+"}", val, -1)
	}
	return a
}

func (t *Testcase) assertStatusHas(matches []string) error {
	expect := t.parseVars(matches[1])
	if !strings.Contains(t.haStatus, expect) {
		return fmt.Errorf("expect status has \"%v\", but got \"%v\"", expect, t.haStatus)
	}
	return nil
}

func (t *Testcase) assertStatusAllOk(matches []string) error {
	if !t.isHaStatusAllOk(t.haStatus) {
		return fmt.Errorf("assert status all ok failed, status was %v", t.haStatus)
	}
	return nil
}

func (t *Testcase) getStatus(matches []string) error {
	ip := t.GetConfigString("remote", "a")
	if output, exitCode, err := cli.Cmdf(ip, t.GetConfigString("remote_path", "status_script")); nil != err {
		return err
	} else if 0 != exitCode {
		return fmt.Errorf("get status return !0 code %v", exitCode)
	} else {
		t.haStatus = ""
		for _, line := range strings.Split(output, "\n") {
			t.haStatus = t.haStatus + "\n" + strings.TrimSpace(line)
		}
	}
	return nil
}

func (t *Testcase) cleanEnv(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}

	if _, _, err := cli.Cmdf(ip, t.GetConfigString("remote_path", "uninstall_script")); nil != err {
		return fmt.Errorf("uninstall HA failed : %v", err)
	}

	time.Sleep(time.Duration(3*t.GetConfigInt("remote", "time_weight")) * time.Second) //wait nssm release (for windows)

	if _, _, err := cli.RmDir(ip, t.GetConfigString("remote_path", "ha_dir")); nil != err {
		return fmt.Errorf("rm actionsky-ha failed : %v", err)
	}

	if _, _, err := cli.RmDir(ip, t.GetConfigString("remote_path", "mysql_install")); nil != err {
		return fmt.Errorf("rm mysql-install failed : %v", err)
	}

	if _, _, err := cli.Rm(ip, t.GetConfigString("remote_path", "ha_zip_file")); nil != err {
		return fmt.Errorf("del HA.zip failed : %v", err)
	}
	return nil
}

func (t *Testcase) parseTime(a string) int {
	var timeout int
	if strings.HasSuffix(a, "min") {
		timeout, _ = strconv.Atoi(strings.TrimSuffix(a, "min"))
		timeout = timeout * 60
	} else if strings.HasSuffix(a, "s") {
		timeout, _ = strconv.Atoi(strings.TrimSuffix(a, "s"))
	}
	return timeout
}

func (t *Testcase) assertNoLog(matches []string) error {
	varName, expect := matches[1], matches[2]
	ip, err := t.getVar(varName)
	if nil != err {
		return err
	}
	found, _, err := cli.FindLog(ip, t.GetConfigString("remote_path", "trace_log"), t.logCheckpoint[ip], expect)
	if nil != err {
		return err
	}
	if found {
		return fmt.Errorf("Should not found log \"%v\", but failed", expect)
	}
	return nil
}

func (t *Testcase) assertLog(matches []string) error {
	varName, expect := matches[1], matches[2]
	timeout := 0
	if len(matches) > 3 {
		timeout = t.parseTime(matches[3])
	}

	ip, err := t.getVar(varName)
	if nil != err {
		return err
	}

	timeWeight := t.GetConfigInt("remote", "time_weight")
	for i := 0; i <= timeout*timeWeight; i++ {
		found, newLogCheckpoint, err := cli.FindLog(ip, t.GetConfigString("remote_path", "trace_log"), t.logCheckpoint[ip], expect)
		if nil != err || !found {
			time.Sleep(1 * time.Second)
			continue
		}
		t.logCheckpoint[ip] = newLogCheckpoint
		return nil
	}
	return fmt.Errorf("No expect log \"%v\" found", expect)
}

func (t *Testcase) vars(matches []string) error {
	varName, ip := matches[1], matches[2]
	t.variables[varName] = ip
	return nil
}

func (t *Testcase) initAutoHaDebug(ip string) error {
	autoQaDebugPath := path.Join(t.GetConfigString("remote_path", "ha_dir"), "auto_qa.debug")
	tempFile := path.Join(os.TempDir(), "auto_qa.debug")
	err := ioutil.WriteFile(tempFile, []byte(""), 0644)
	if nil != err {
		return err
	}
	defer os.Remove(tempFile)
	if err := cli.PushFile(ip, tempFile, autoQaDebugPath); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) runHaInstallScript(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if _, _, err := cli.Cmdf(ip, t.GetConfigString("remote_path", "install_script")); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) extractHaInstallation(matches []string) error {
	ip, err := t.getVar(matches[1])
	if nil != err {
		return err
	}
	if t.isWindows() {
		if err := cli.PushFile(ip, "./7za.exe", "c:\\7za.exe"); nil != err {
			return err
		}
	}
	if err := cli.PushFile(ip, t.haZip, t.GetConfigString("remote_path", "ha_zip_file")); nil != err {
		return err
	}
	time.Sleep(time.Duration(3*t.GetConfigInt("remote", "time_weight")) * time.Second) //wait write file done (for windows)
	if _, _, err := cli.Cmdf(ip, t.GetConfigString("remote_command", "unzip_ha_installation")); nil != err {
		return err
	}
	if err := cli.PushFile(ip, t.GetConfigString("local_path", "ha_config"), t.GetConfigString("remote_path", "ha_config")); nil != err {
		return err
	}
	time.Sleep(time.Duration(1*t.GetConfigInt("remote", "time_weight")) * time.Second) //wait write file done (for windows)
	if err := t.initAutoHaDebug(ip); nil != err {
		return err
	}
	return nil
}

func (t *Testcase) installHa(matches []string) error {
	if err := t.doAction(fmt.Sprintf("extract HA installation on {%v}", matches[1])); nil != err {
		return err
	}
	if err := t.doAction(fmt.Sprintf("run HA install script on {%v}", matches[1])); nil != err {
		return err
	}

	return nil
}

func (t *Testcase) getVar(varname string) (string, error) {
	if ret, ok := t.variables[varname]; ok {
		return ret, nil
	} else {
		return "", fmt.Errorf("no variable found :%v", varname)
	}
}

func (t *Testcase) onFailure() {
	for _, ip := range []string{t.GetConfigString("remote", "a"), t.GetConfigString("remote", "b")} {
		cli.PullFile(ip, t.GetConfigString("remote_path", "ha_log"), fmt.Sprintf("%v/%v_ha.log", t.logDir, ip))
		cli.PullFile(ip, t.GetConfigString("remote_path", "trace_log"), fmt.Sprintf("%v/%v_trace.log", t.logDir, ip))
		cli.PullFile(ip, t.GetConfigString("remote_path", "target_service_log"), fmt.Sprintf("%v/%v_target_service.log", t.logDir, ip))
	}
}
