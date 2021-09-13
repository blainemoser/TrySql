package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/blainemoser/TrySql/help"
	"github.com/blainemoser/TrySql/trysql"
	"github.com/blainemoser/TrySql/utils"
)

type TestSuiteTS struct {
	Shell *TrySqlShell
	TS    *trysql.TrySql
}

func (ts *TestSuiteTS) HandlePanic() {
	r := recover()
	if r != nil {
		if ts.Shell != nil && len(ts.Shell.ShellOutChan) > 0 {
			<-ts.Shell.ShellOutChan
		}
		panic(r)
	}
}

func InitialiseTestSuite() (*TestSuiteTS, error) {
	trysql.Testing = true
	Testing = true
	trySql := trysql.Initialise()
	ts := &TestSuiteTS{
		Shell: New(trySql),
		TS:    trySql,
	}
	return ts, nil
}

func (ts *TestSuiteTS) Start() {
	ts.Shell.Start(true)
}

func (ts *TestSuiteTS) Stop() error {
	err := ts.TS.TearDown()
	if err != nil {
		return err
	}
	trysql.Testing = false
	return nil
}

func (ts *TestSuiteTS) sendSignal(funcCall interface{}, signal string) chan bool {
	defer ts.HandlePanic()
	waitChan := make(chan bool, 1)
	ts.Shell.Push(signal)
	<-ts.Shell.ShellOutChan
	result := ts.Shell.LastOutput()
	ts.check(funcCall, waitChan, result)
	return waitChan
}

func (ts *TestSuiteTS) check(funcInterface interface{}, waitChan chan bool, check string) {
	functionCall, ok := (funcInterface).(func(chan bool, string))
	if !ok {
		panic(fmt.Errorf("invalid function provided"))
	}
	functionCall(waitChan, check)
}

func (ts *TestSuiteTS) SendHelpSignal() chan bool {
	return ts.sendSignal(ts.checkHelp, "help")
}

func (ts *TestSuiteTS) SendVersionSignal() chan bool {
	return ts.sendSignal(ts.checkVersion, "version")
}

func (ts *TestSuiteTS) SendHistorySignal() {
	<-ts.SendVersionSignal()
	ts.Shell.Push("history")
	ts.checkHistory()
}

func (ts *TestSuiteTS) SendContainerDetailsSignal() chan bool {
	return ts.sendSignal(ts.checkDetails, "cd")
}

func (ts *TestSuiteTS) SendContainerIDSignal() chan bool {
	return ts.sendSignal(ts.checkID, "cid")
}

func (ts *TestSuiteTS) SendTempPassSignal() chan bool {
	return ts.sendSignal(ts.checkTempPass, "temp-password")
}

func (ts *TestSuiteTS) SendPassSignal() chan bool {
	return ts.sendSignal(ts.checkPass, "password")
}

func (ts *TestSuiteTS) SendQuerySignal() chan bool {
	return ts.sendSignal(ts.checkQuery, "query SHOW VARIABLES LIKE 'max_connections'")
}

func (ts *TestSuiteTS) SendExitSignal() {
	ts.Shell.OsInterrupt <- os.Interrupt
}

func (ts *TestSuiteTS) IncrementWG() {
	ts.Shell.WG.Add(1)
}

func (ts *TestSuiteTS) DecrementWG() {
	ts.Shell.WG.Done()
}

func (ts *TestSuiteTS) checkHelp(waitChan chan bool, output string) {
	help := help.Get([]string{"help"})
	helpSplit := strings.Split(help, "\n\n")
	var errs []error
	for _, h := range helpSplit {
		h = strings.Trim(h, " ")
		if len(h) < 1 {
			continue
		}
		if !strings.Contains(output, h) {
			errs = append(errs, fmt.Errorf("expected output to contain '%s'", h))
		}
	}
	var err error
	if len(errs) > 0 {
		errs = append(errs, fmt.Errorf("output was '%s'", output))
		err = utils.GetErrors(errs)
	}
	if err != nil {
		panic(err)
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkVersion(waitChan chan bool, output string) {
	version := ts.TS.DockerVersion()
	if !strings.Contains(output, version) {
		panic(fmt.Errorf("expected output to contain '%s'", version))
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkTempPass(waitChan chan bool, output string) {
	password := ts.TS.DockerTempPassword()
	if len(password) < 1 {
		panic(fmt.Errorf("temp password not set"))
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkPass(waitChan chan bool, output string) {
	password := ts.TS.CurrentPassword()
	if len(password) < 1 {
		panic(fmt.Errorf("password not set"))
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkQuery(waitChan chan bool, output string) {
	if !strings.Contains(strings.ReplaceAll(output, "\n", " "), "Variable_name") {
		panic(fmt.Errorf("expected output to contain 'Variable_name'"))
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkHistory() {
	if !strings.Contains(strings.ToLower(TestHistoryOutput), "docker version") {
		panic(fmt.Errorf("expected output to contain 'docker version'"))
	}
}

func (ts *TestSuiteTS) checkDetails(waitChan chan bool, output string) {
	if !strings.Contains(strings.ToLower(output), "trysql") {
		panic(fmt.Errorf("expected output to contain the container name 'TrySql', got '%s'", output))
	}
	waitChan <- true
}

func (ts *TestSuiteTS) checkID(waitChan chan bool, output string) {
	expects := map[string]bool{
		"not found": true,
		"something went wrong while trying to get the container's details": true,
	}
	if expects[output] {
		panic(fmt.Errorf(output))
	}
	waitChan <- true
}
