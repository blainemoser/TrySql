package main

import (
	"os"
	"testing"

	test "github.com/blainemoser/TrySql/trysqltest"
	"github.com/blainemoser/TrySql/utils"
)

var suite *test.TestSuiteTS

func TestMain(m *testing.M) {
	var err error
	suite, err = test.Init()
	if err != nil {
		panic(err)
	}
	suite.Start()
	code := m.Run()
	err = suite.Stop()
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestHelp(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendHelpSignal()
}

func TestVersion(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendVersionSignal()
}

func TestHistory(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendHistorySignal()
}
