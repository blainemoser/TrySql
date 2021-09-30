package trysql

import (
	"fmt"
	"os"
	"strings"
	"testing"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/TrySql/utils"
)

var tsql *TrySql
var torn bool
var initialised bool

func TestMain(m *testing.M) {
	Testing = true
	code := m.Run()
	Testing = false
	if !torn {
		err := tsql.TearDown()
		if err != nil {
			panic(err)
		}
	}
	os.Exit(code)
}

func TestInitialize(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
}

func TestListContainers(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result, err := tsql.listContainers(false)
	if err != nil {
		t.Error(err)
	}
	expects := "mysql/mysql-server:latest"
	for _, container := range result {
		if strings.Contains(container, expects) {
			return
		}
	}
	t.Error(fmt.Errorf("expected to find container for '%s'", expects))
}

func TestDockerVersion(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.DockerVersion()
	if !strings.Contains(strings.ToLower(result), "docker version") {
		t.Error("expected to find docker version")
	}
	if !strings.Contains(strings.ToLower(result), "build") {
		t.Error("expected to find docker version build")
	}
}

func TestPassword(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.Password()
	if len(result) != 32 {
		t.Errorf("expected password to be 32 characters, got %d", len(result))
	}
}

func TestMySQLCommand(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.MySQLCommand()
	expects := fmt.Sprintf(
		"mysql -uroot -p%s -h127.0.0.1 -P%s",
		tsql.Password(),
		tsql.HostPortStr(),
	)
	if result != expects {
		t.Errorf("expected command to be %s, got %s", expects, result)
	}
}

func TestGetContainerDetails(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.GetContainerDetails(false)
	if !strings.Contains(strings.ToLower(result), "trysql") {
		t.Errorf("expected to find the container name 'TrySql', got '%s'", result)
	}
	result = tsql.GetContainerDetails(true)
	expects := map[string]bool{
		"not found": true,
		"something went wrong while trying to get the container's details": true,
	}
	if expects[result] {
		t.Errorf(result)
	}
}

func TestMysqlArgs(t *testing.T) {
	defer utils.HandelPanic(t)
	result := tsql.mysqlArgs("QUERY")
	if !strings.Contains(result, "--execute=\"QUERY\"") {
		t.Errorf("expected raw string to contain '--execute=\"QUERY\"', got %s", result)
	}
}

func TestQuery(t *testing.T) {
	defer utils.HandelPanic(t)
	result, err := tsql.Query("SHOW VARIABLES LIKE 'max_connections'", true)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(result, "Variable_name") {
		t.Errorf("expected result to contain 'Variable_name', got '%s'", result)
	}
}

func TestDetails(t *testing.T) {
	defer utils.HandelPanic(t)
	result := tsql.GetDetails([]string{"details", "Id", "State/Health/Log"})
	if !strings.Contains(result, "Id:") {
		t.Errorf("expected result to contain 'Id:', got '%s'", result)
	}
	if !strings.Contains(result, "State/Health/Log:") {
		t.Errorf("expected result to contain 'State/Health/Log:', got '%s'", result)
	}
	log := strings.Split(result, "State/Health/Log:")
	if len(log) < 2 {
		t.Errorf("expected log output in the results")
	}
	logString := strings.Trim(strings.Trim(log[1], "\n"), " ")
	js := &jsonextract.JSONExtract{
		RawJSON: logString,
	}
	_, err := js.Extract("[0]")
	if err != nil {
		t.Errorf("expected at least one log: %s", err.Error())
	}
}

func TestContainerRunning(t *testing.T) {
	defer utils.HandelPanic(t)
	result, err := tsql.containerRunning()
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Errorf("expected container to be running")
	}
}

func TestTearDown(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	err := tsql.TearDown()
	if err != nil {
		t.Error(err)
	}
	torn = true
}

func tInit() {
	var err error
	if !initialised {
		tsql, err = Initialise([]string{})
		if err != nil {
			panic(err)
		}
		initialised = true
	}
}
