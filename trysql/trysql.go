package trysql

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blainemoser/TrySql/configs"
	"github.com/blainemoser/TrySql/docker"
	"github.com/blainemoser/TrySql/utils"
	"github.com/gosuri/uilive"
)

var Testing bool

type TrySql struct {
	docker     *docker.Docker
	image      string
	hash       string
	ReadyState int
	Configs    *configs.Configs
}

func Initialise() *TrySql {
	var err error
	args := getArgs()
	confs, err := configs.New(args)
	if err != nil {
		log.Fatal(err.Error())
	}
	ts, err := generate(utils.GetProcessOwner(), confs)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("found " + string(ts.DockerVersion()))
	err = ts.provision()
	if err != nil {
		log.Fatal("could not provision container - " + err.Error())
	}
	err = ts.run()
	if err != nil {
		log.Fatal("could not run container - " + err.Error())
	}
	err = ts.waitForHealthy()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ts.setNewPass()
	if err != nil {
		tdErr := ts.TearDown()
		if tdErr != nil {
			log.Fatal(tdErr)
		}
		log.Fatal(err.Error())
	}
	return ts
}

func generate(owner string, configs *configs.Configs) (*TrySql, error) {
	ts := &TrySql{
		docker: &docker.Docker{
			RunAsSudo: owner != "root",
		},
		image:   "mysql/mysql-server:" + configs.GetMysqlVersion(),
		Configs: configs,
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

func (ts *TrySql) DockerTempPassword() string {
	return ts.docker.GeneratedRootPassword
}

func (ts *TrySql) Query(query string, report bool) (string, error) {
	result, err := ts.outputCommandRaw(ts.mysqlArgs(query))
	if err != nil {
		errString := strings.ReplaceAll(err.Error(), "\n", " | ")
		if strings.Contains(err.Error(), "ERROR") {
			return result, errors.New(errString)
		}
		if report {
			result = result + " | " + errString
		}
	}
	return result, nil
}

func (ts *TrySql) TearDown() error {
	running, err := ts.containerRunning()
	if !running {
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Println("tearing down")
	err = ts.waitAndWrite(ts.stoppingContainer, "stopping container")
	if err != nil {
		return err
	}
	return ts.waitAndWrite(ts.removingContainer, "removing container")
}

func (ts *TrySql) CurrentPassword() string {
	if ts.docker.CurrentPassword == "" {
		return ts.DockerTempPassword()
	}
	return ts.docker.CurrentPassword
}

func (ts *TrySql) mysqlArgs(query string) string {
	return fmt.Sprintf(
		"exec TrySql mysql --user=root --password=\"%s\" --execute=\"%s\" --connect-expired-password",
		ts.CurrentPassword(),
		query,
	)
}

func (ts *TrySql) GetContainerDetails(idOnly bool) string {
	containers, err := ts.ps()
	if err != nil {
		return "something went wrong while trying to get the container's details"
	}
	result, err := ts.findContainer(containers)
	if err != nil {
		return err.Error()
	}
	if !idOnly {
		return result
	}
	ts.filterContainerID(&result)
	return result
}

func (ts *TrySql) setHealthyStatus() error {
	timeout := 0
	wait := time.NewTicker(time.Second)
	status := make(chan bool, 1)
	errLog := make(chan error, 1)
	ts.getHealthStatus(status, errLog)
	for {
		select {
		case err := <-errLog:
			return err
		case <-status:
			return nil
		case <-wait.C:
			timeout += 1
			ts.getHealthStatus(status, errLog)
			if timeout >= 120 {
				return fmt.Errorf("timed out while waiting for container temporary password")
			}
		}
	}
}

func (ts *TrySql) getHealthStatus(status chan bool, errorChan chan error) {
	details := ts.GetContainerDetails(false)
	details = strings.ToLower(details)
	if strings.Contains(details, "(health: starting)") {
		return
	}
	if strings.Contains(details, "(healthy)") {
		status <- true
		return
	}
	errorChan <- errors.New("no startup activity on container")
}

func (ts *TrySql) setNewPass() error {
	err := ts.getPassLog()
	if err != nil {
		return err
	}
	newPass, _ := utils.MakePass()
	result, err := ts.Query("ALTER USER 'root'@'localhost' IDENTIFIED BY '"+newPass+"';", true)
	fmt.Println(result)
	if err != nil {
		return err
	}
	ts.docker.CurrentPassword = newPass
	return nil
}

func (ts *TrySql) getPassLog() error {
	logs, err := ts.getLogs()
	if err != nil {
		return err
	}
	for _, log := range logs {
		if strings.Contains(log, "GENERATED ROOT PASSWORD") {
			ts.docker.GeneratedRootPassword = ts.extractPassword(log)
			return nil
		}
	}
	return errors.New("temp password not found")
}

func (ts *TrySql) extractPassword(log string) string {
	log = strings.Replace(log, "[Entrypoint]", "", 1)
	log = strings.Replace(log, "GENERATED ROOT PASSWORD:", "", 1)
	log = strings.Trim(log, " ")
	return log
}

func (ts *TrySql) getLogs() ([]string, error) {
	result, err := ts.outputCommand([]string{"logs", "TrySql"})
	if err != nil {
		return nil, err
	}
	return strings.Split(result, "\n"), nil
}

func (ts *TrySql) listContainers(all bool) ([]string, error) {
	args := []string{"container", "ls"}
	if all {
		args = append(args, "-al")
	}
	containers, err := ts.outputCommand(args)
	if err != nil {
		return nil, err
	}
	existingContainers := strings.Split(containers, "\n")
	if len(existingContainers) > 0 {
		if strings.Contains(existingContainers[0], "CONTAINER ID") {
			existingContainers = existingContainers[1:]
		}
		return existingContainers, nil
	}

	return []string{}, nil
}

func (ts *TrySql) ps() ([]string, error) {
	containers, err := ts.outputCommand([]string{"ps"})
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

func (ts *TrySql) provision() error {
	msg := "pulling up to date image"
	return ts.waitAndWrite(ts.provisioningDocker, msg)
}

func (ts *TrySql) waitForHealthy() error {
	msg := "waiting for container to set up"
	return ts.waitAndWrite(ts.waitingForHealtyStatus, msg)
}

func (ts *TrySql) containerRunning() (bool, error) {
	exists, err := ts.containerExists(false)
	if err != nil {
		return false, err
	}
	if !exists {
		fmt.Println("container does not exist")
		return false, nil
	}
	running, err := ts.isRunning()
	if err != nil {
		return false, err
	}
	if !running {
		fmt.Println("container is not running")
		return false, nil
	}
	return true, nil
}

func (ts *TrySql) run() error {
	running, err := ts.containerExists(false)
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

func (ts *TrySql) findContainer(containers []string) (string, error) {
	for _, container := range containers {
		if strings.Contains(container, "TrySql") {
			return container, nil
		}
	}
	return "", errors.New("not found")
}

func (ts *TrySql) containerExists(all bool) (bool, error) {
	containers, err := ts.listContainers(all)
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

func (ts *TrySql) filterContainerID(containerDetails *string) {
	*containerDetails = strings.ReplaceAll(*containerDetails, "\n", " ")
	*containerDetails = strings.ReplaceAll(*containerDetails, "\t", " ")
	splitC := strings.Split(*containerDetails, " ")
	if len(splitC) < 1 {
		*containerDetails = ""
	}
	*containerDetails = splitC[0]
}

func (ts *TrySql) isRunning() (bool, error) {
	containers, err := ts.ps()
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
	err := ts.needsCleanup()
	if err != nil {
		initChan <- err
		return
	}
	result, err := ts.outputCommand([]string{"run", "-d", "-e", "3306", "-p", "6603:3306", "--name=TrySql", ts.image})
	ts.hash = result
	ts.ReadyState = 1
	initChan <- err
}

func (ts *TrySql) needsCleanup() error {
	exists, err := ts.containerExists(true)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return ts.cleanUp()
}

func (ts *TrySql) cleanUp() error {
	ts.outputCommand([]string{"stop", "TrySql"})
	_, err := ts.outputCommand([]string{"rm", "TrySql"})
	return err
}

func (ts *TrySql) provisioningDocker(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	_, err := ts.outputCommand([]string{"pull", ts.image})
	initChan <- err
}

func (ts *TrySql) waitingForHealtyStatus(wg *sync.WaitGroup, initChan chan error) {
	defer wg.Done()
	initChan <- ts.setHealthyStatus()
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

func (ts *TrySql) outputCommandRaw(arg string) (string, error) {
	result, err := ts.docker.Com().ExecRaw(arg)
	if err != nil {
		return "", err
	}
	return result, nil
}

func getArgs() []string {
	if Testing {
		return []string{"-v", "latest"}
	}
	return os.Args[1:]
}
