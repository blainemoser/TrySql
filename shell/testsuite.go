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
	ts.Shell.Push("help")
	ts.checkHelp(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendVersionSignal() {
	ts.Shell.Push("version")
	ts.checkVersion(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendHistorySignal() {
	ts.Shell.Push("version")
	ts.Shell.Push("history")
	ts.checkHistory(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendExitSignal() {
	ts.Shell.OsInterrupt <- os.Interrupt
}

func (ts *TestSuiteTS) SendQuit() {
	ts.Shell.handleCommand("version")
	fmt.Println(ts.Shell.LastOutput())
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
		panic(fmt.Errorf("expected output to containt '%s'", version))
	}
}

func (ts *TestSuiteTS) checkHistory(output string) {
	if !strings.Contains(strings.ToLower(output), "docker version") {
		panic(fmt.Errorf("expected output to contain 'docker version'"))
	}
}