package shell

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blainemoser/TrySql/help"
	"github.com/blainemoser/TrySql/trysql"
	"github.com/blainemoser/TrySql/utils"
)

type TestSuiteTS struct {
	Shell *TrySqlShell
	TS    *trysql.TrySql
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

func (ts *TestSuiteTS) SendHelpSignal() {
	result := ts.Shell.Push("help")
	ts.checkHelp(result)
}

func (ts *TestSuiteTS) SendVersionSignal() {
	result := ts.Shell.Push("version")
	ts.checkVersion(result)
}

func (ts *TestSuiteTS) SendHistorySignal() {
	result := ts.Shell.Push("version")
	ts.Shell.Push("history")
	ts.checkHistory(result)
}

func (ts *TestSuiteTS) SendContainerDetailsSignal() {
	result := ts.Shell.Push("cd")
	fmt.Println(result)
	time.Sleep(time.Second * 1)
	result = ts.Shell.LastOutput()
	ts.checkDetails(result)
}

func (ts *TestSuiteTS) SendContainerIDSignal() {
	result := ts.Shell.Push("cid")
	ts.checkID(result)
}

func (ts *TestSuiteTS) SendTempPassSignal() {
	result := ts.Shell.Push("temp-password")
	ts.checkTempPass(result)
}

func (ts *TestSuiteTS) SendExitSignal() {
	ts.Shell.OsInterrupt <- os.Interrupt
}

func (ts *TestSuiteTS) SendQuit() {
	result := ts.Shell.handleCommand("version")
	fmt.Println(result)
}

func (ts *TestSuiteTS) IncrementWG() {
	ts.Shell.WG.Add(1)
}

func (ts *TestSuiteTS) DecrementWG() {
	ts.Shell.WG.Done()
}

func (ts *TestSuiteTS) checkHelp(output string) {
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
}

func (ts *TestSuiteTS) checkVersion(output string) {
	version := ts.TS.DockerVersion()
	if !strings.Contains(output, version) {
		panic(fmt.Errorf("expected output to contain '%s'", version))
	}
}

func (ts *TestSuiteTS) checkTempPass(output string) {
	password := ts.TS.DockerTempPassword()
	if len(password) < 1 {
		panic(fmt.Errorf("temp password not set"))
	}
}

func (ts *TestSuiteTS) checkHistory(output string) {
	if !strings.Contains(strings.ToLower(output), "docker version") {
		panic(fmt.Errorf("expected output to contain 'docker version'"))
	}
}

func (ts *TestSuiteTS) checkDetails(output string) {
	if !strings.Contains(strings.ToLower(output), "trysql") {
		panic(fmt.Errorf("expected output to contain the container name 'TrySql', got '%s'", output))
	}
}

func (ts *TestSuiteTS) checkID(output string) {
	expects := map[string]bool{
		"not found": true,
		"something went wrong while trying to get the container's details": true,
	}
	if expects[output] {
		panic(fmt.Errorf(output))
	}
}
