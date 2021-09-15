package docker

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/blainemoser/TrySql/configs"
	"github.com/blainemoser/TrySql/utils"
)

type Docker struct {
	Version   string
	Password  string
	HostPort  int
	RunAsSudo bool
}

type command struct {
	inputs []string
	d      *Docker
}

func New(configs *configs.Configs) *Docker {
	owner := utils.GetProcessOwner()
	password, _ := utils.MakePass()
	return &Docker{
		RunAsSudo: owner != "root",
		Password:  password,
		HostPort:  configs.GetPort(),
	}
}

func (d *Docker) Com() *command {
	var inputs []string
	if d.RunAsSudo {
		inputs = []string{"sudo", "docker"}
	} else {
		inputs = []string{"docker"}
	}
	com := &command{
		inputs: inputs,
		d:      d,
	}
	return com
}

func (d *Docker) SetVersion() error {
	result, err := d.Com().Args([]string{"-v"}).Exec()
	if err != nil {
		return err
	}
	d.Version = strings.Replace(result, "\n", "", -1)
	return nil
}

func (c *command) Args(args []string) *command {
	c.inputs = append(c.inputs, args...)
	return c
}

func (c *command) Exec() (string, error) {
	result, err := exec.Command(c.inputs[0], c.inputs[1:]...).CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + ": " + string(result))
	}
	return string(result), nil
}

func (c *command) ExecRaw(arg string) (string, error) {
	var result []byte
	var err error
	if c.d.RunAsSudo {
		result, err = exec.Command("sh", "-c", c.inputs[0]+" "+c.inputs[1]+" "+arg).CombinedOutput()
	} else {
		result, err = exec.Command("sh", "-c", c.inputs[0]+" "+arg).CombinedOutput()
	}
	if err != nil {
		return "", errors.New(err.Error() + ": " + string(result))
	}
	return string(result), nil
}
