package docker

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/blainemoser/TrySql/utils"
)

var docker *Docker

func TestMain(m *testing.M) {
	docker = &Docker{
		RunAsSudo: utils.GetProcessOwner() != "root",
	}
	code := m.Run()
	os.Exit(code)
}

func TestCom(t *testing.T) {
	com := docker.Com()
	hasDocker := false
	var unexp []string
	for _, input := range com.inputs {
		if input == "docker" {
			hasDocker = true
		}
		if input != "docker" && input != "sudo" {
			unexp = append(unexp, input)
		}
	}
	var errs []error
	if !hasDocker {
		errs = append(errs, fmt.Errorf("expected command input to contain 'docker'"))
	}
	if len(unexp) > 0 {
		errs = append(errs, fmt.Errorf("found unexpected inputs: %s", strings.Join(unexp, "; ")))
	}
	err := utils.GetErrors(errs)
	if err != nil {
		t.Error(err)
	}
}

func TestSetVersion(t *testing.T) {
	docker.SetVersion()
	if len(docker.Version) < 1 {
		t.Errorf("docker version not set")
	}
	if !strings.Contains(strings.ToLower(docker.Version), "docker version") {
		t.Errorf("expected docker version to report a docker version")
	}
}

func TestArgs(t *testing.T) {
	com := docker.Com().Args([]string{"new-arg"})
	for _, arg := range com.inputs {
		if arg == "new-arg" {
			return
		}
	}
	t.Errorf("expected command to contain 'new-arg'")
}

func TestExec(t *testing.T) {
	result, err := docker.Com().Args([]string{"container", "ls"}).Exec()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(result, "CONTAINER ID") {
		t.Errorf("expected result to contain a container id header")
	}
}

func TestExecRaw(t *testing.T) {
	result, err := docker.Com().ExecRaw("container ls -al")
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(result, "CONTAINER ID") {
		t.Errorf("expected result to contain a container id header")
	}
}
