package docker

import (
	"errors"
	"os/exec"
	"strings"
)

type Docker struct {
	Version   string
	RunAsSudo bool
}

type command struct {
	inputs []string
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
