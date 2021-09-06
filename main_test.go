package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/blainemoser/TrySql/test"
)

var suite *test.TestSuiteTS

func TestMain(m *testing.M) {
	var err error
	suite, err = test.Init()
	if err != nil {
		panic(err)
	}
	suite.Shell.WG.Add(1)
	suite.Shell.StartTest()
	code := m.Run()
	suite.Shell.WG.Wait()
	err = suite.Stop()
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestProg(t *testing.T) {
	defer suite.HandelPanic()
	helpResult := suite.SendHelpSignal()
	fmt.Println(helpResult)
	historyResult := suite.SendHistorySignal()
	fmt.Println(historyResult)
	suite.SendExitSignal()
}
