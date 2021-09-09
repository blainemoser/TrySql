package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/blainemoser/TrySql/help"
	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/trysql"
	"github.com/blainemoser/TrySql/utils"
)

type TestSuiteTS struct {
	Shell *shell.TrySqlShell
	TS    *trysql.TrySql
}

func Init() (*TestSuiteTS, error) {
	trysql.Testing = true
	shell.Testing = true
	trySql := trysql.Initialise()
	ts := &TestSuiteTS{
		Shell: shell.New(trySql),
		TS:    trySql,
	}

	err := ts.run()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (ts *TestSuiteTS) Start() {
	ts.Shell.StartTest()
}

func (ts *TestSuiteTS) Stop() error {
	err := ts.TS.TearDown()
	if err != nil {
		return err
	}
	trysql.Testing = false
	return nil
}

func (ts *TestSuiteTS) HandelPanic(t *testing.T) {
	r := recover()
	if r != nil {
		t.Error(r)
	}
}

func (ts *TestSuiteTS) SendHelpSignal() {
	ts.Shell.Push("help")
	checkHelp(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendVersionSignal() {
	ts.Shell.Push("version")
	checkVersion(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendHistorySignal() {
	ts.Shell.Push("history")
	checkHistory(ts.Shell.LastOutput())
}

func (ts *TestSuiteTS) SendExitSignal() {
	ts.Shell.OsInterrupt <- os.Interrupt
}

func (ts *TestSuiteTS) run() error {
	err := ts.TS.Run()
	if err != nil {
		return err
	}
	return nil
}

func checkHelp(output string) {
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
	err := utils.GetErrors(errs)
	if err != nil {
		panic(err)
	}
}

func checkVersion(output string) {
	fmt.Println(output)
}

func checkHistory(output string) {
	fmt.Println(output)
}
