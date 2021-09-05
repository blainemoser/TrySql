package trysql

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blainemoser/TrySql/docker"
	"github.com/gosuri/uilive"
)

type TrySql struct {
	docker *docker.Docker
	image  string
	hash   string
}

func Init(owner, version string) (*TrySql, error) {
	ts := &TrySql{
		docker: &docker.Docker{
			RunAsSudo: owner != "root",
		},
		image: "mysql/mysql-server:" + version,
	}
	err := ts.initDocker()
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func (ts *TrySql) DockerVersion() string {
	return ts.docker.Version
}

func (ts *TrySql) ListContainers() ([]string, error) {
	containers, err := ts.outputCommand([]string{"container", "ls"})
	if err != nil {
		return nil, err
	}
	runningContainers := strings.Split(containers, "\n")
	if len(runningContainers) > 0 {
		if strings.Contains(runningContainers[0], "CONTAINER ID") {
			runningContainers = runningContainers[1:]
		}
		return runningContainers, nil
	}

	return []string{}, nil
}

func (ts *TrySql) Provision() error {
	msg := "pulling up to date image"
	return ts.waitAndWrite(ts.provisioningDocker, msg)
}

func (ts *TrySql) TearDown() error {
	running, err := ts.running()
	if err != nil {
		return err
	}
	if !running {
		fmt.Println("container not running")
		return nil
	}
	err = ts.waitAndWrite(ts.stoppingContainer, "stopping container")
	if err != nil {
		return err
	}
	return ts.waitAndWrite(ts.removingContainer, "removing container")
}

func (ts *TrySql) Run() error {
	running, err := ts.running()
	if err != nil {
		return err
	}
	if running {
		return nil
	}
	return ts.runNew()
}

func (ts *TrySql) waitAndWrite(funcInterface interface{}, msg string) error {
	functionCall, ok := (funcInterface).(func(*sync.WaitGroup, chan error))
	if !ok {
		return fmt.Errorf("invalid function provided")
	}
	var err error
	initChan := make(chan error, 1)
	writer := uilive.New() // writer for the first line
	wg := &sync.WaitGroup{}
	writer.Start()
	wg.Add(2)
	go ts.wait(wg, initChan, writer, &err, msg)
	go functionCall(wg, initChan)
	wg.Wait()
	close(initChan)
	fmt.Fprintf(writer, msg+" %s\n", "done")
	writer.Stop()
	return err
}

func (ts *TrySql) initDocker() error {
	return ts.docker.SetVersion()
}

func (ts *TrySql) running() (bool, error) {
	containers, err := ts.ListContainers()
	if err != nil {
		return false, err
	}
	for _, container := range containers {
		if strings.Contains(container, "TrySql") {
			return true, nil
		}
	}
	return false, nil
}

func (ts *TrySql) runNew() error {
	msg := "waiting for container connection"
	return ts.waitAndWrite(ts.settingUpContainer, msg)
}

func (ts *TrySql) wait(wg *sync.WaitGroup, initChan chan error, writer *uilive.Writer, err *error, msg string) {
	timeOut := 0
	updating := []string{"|", "/", "-", "\\"}
	uIndex := 0
	for {
		select {
		case *err = <-initChan:
			wg.Done()
			return
		default:
			fmt.Fprintf(writer, msg+" %s\n", updating[uIndex])
			if timeOut > 900 {
				*err = fmt.Errorf("timed out")
				initChan <- *err
				wg.Done()
				return
			}
			if uIndex == 3 {
				uIndex = 0
			} else {
				uIndex++
			}
			timeOut++
			time.Sleep(time.Millisecond * 200)
		}
	}
}

func (ts *TrySql) settingUpContainer(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	result, err := ts.outputCommand([]string{"run", "-d", "--name=TrySql", ts.image})
	ts.hash = result
	initChan <- err
}

func (ts *TrySql) provisioningDocker(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	_, err := ts.outputCommand([]string{"pull", ts.image})
	initChan <- err
}

func (ts *TrySql) stoppingContainer(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	_, err := ts.outputCommand([]string{"container", "stop", "TrySql"})
	initChan <- err
}

func (ts *TrySql) removingContainer(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	_, err := ts.outputCommand([]string{"container", "rm", "TrySql"})
	initChan <- err
}

func (ts *TrySql) outputCommand(args []string) (string, error) {
	result, err := ts.docker.Com().Args(args).Exec()
	if err != nil {
		return "", err
	}
	return result, nil
}