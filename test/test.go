package test

import (
	"fmt"
	"os"

	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/trysql"
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
		Shell: shell.New(trySql.Configs),
		TS:    trySql,
	}

	err := ts.start()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (ts *TestSuiteTS) start() error {
	err := ts.TS.Run()
	if err != nil {
		return err
	}
	return nil
}

func (ts *TestSuiteTS) Stop() error {
	err := ts.TS.TearDown()
	if err != nil {
		return err
	}
	trysql.Testing = false
	return nil
}

func (ts *TestSuiteTS) HandelPanic() {
	r := recover()
	if r != nil {
		if err, ok := r.(error); ok {
			fmt.Printf("recovering from panic: %s\n", err.Error())
		}
		tdErr := ts.TS.TearDown()
		if tdErr != nil {
			fmt.Printf("error while tearing down test: %s", tdErr.Error())
		}
		panic(r)
	}
}

func (ts *TestSuiteTS) SendHelpSignal() string {
	ts.Shell.Push("help")
	return ts.Shell.LastOutput()
}

func (ts *TestSuiteTS) SendHistorySignal() string {
	ts.Shell.Push("history")
	return ts.Shell.LastOutput()
}

func (ts *TestSuiteTS) SendExitSignal() {
	ts.Shell.OsInterrupt <- os.Interrupt
}
