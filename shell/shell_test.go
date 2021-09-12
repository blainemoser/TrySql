package shell

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/blainemoser/TrySql/utils"
)

var suite *TestSuiteTS

func TestMain(m *testing.M) {
	var err error
	suite, err = InitialiseTestSuite()
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

func TestNew(t *testing.T) {
	defer utils.HandelPanic(t)
}

func TestLastCommand(t *testing.T) {
	defer utils.HandelPanic(t)
	lastOutput := suite.Shell.Push("version")
	if !strings.Contains(strings.ToLower(lastOutput), "docker version") {
		t.Errorf("expected last command to contain 'docker version', got '%s'", lastOutput)
	}
}

func TestLastOutput(t *testing.T) {
	defer utils.HandelPanic(t)
	lastOutput := suite.Shell.Push("help version")
	expects := "Gets the Docker Version"
	if !strings.Contains(lastOutput, expects) {
		t.Errorf("expected output to contain '%s', got '%s'", expects, lastOutput)
	}
}

func TestSanitize(t *testing.T) {
	defer utils.HandelPanic(t)
	input := `     here is 
		bad input
	`
	suite.Shell.sanitize(&input)
	expects := "here is bad input"
	if input != expects {
		t.Error(fmt.Errorf("expected output to be '%s', got '%s'", expects, input))
	}
}

func TestSpecial(t *testing.T) {
	suite.Shell.Push("version")
	testString := string([]byte{27, 91, 65, 10})
	suite.Shell.special(&testString)
	if testString != "version" {
		t.Errorf("expected input to have been changed to last command 'version', got '%s'", testString)
	}
	testString = string([]byte{27, 91, 50, 10})
	suite.Shell.special(&testString)
	if testString != "" {
		t.Errorf("expected input to have been changed empty string, got '%s'", testString)
	}
}

func TestTempPass(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendTempPassSignal()
}

func TestHelp(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendHelpSignal()
}

func TestHistory(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendHistorySignal()
}

func TestContainerDetails(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendContainerDetailsSignal()
}

func TestContainerID(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.SendContainerIDSignal()
}

func TestQuit(t *testing.T) {
	defer utils.HandelPanic(t)
	suite.IncrementWG()
	suite.SendExitSignal()
}

func TestStart(t *testing.T) {
	defer utils.HandelPanic(t)
}
