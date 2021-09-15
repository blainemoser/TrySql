package configs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blainemoser/TrySql/utils"
)

type Configs struct {
	inputs       map[string][]string
	MysqlVersion string
	BufferSize   int
}

func New(inputs []string) (*Configs, error) {
	parsed, err := setInputs(inputs)
	if err != nil {
		return nil, err
	}
	return &Configs{
		inputs: parsed,
	}, nil
}

func expected() map[string]string {
	return map[string]string{
		"v":           "MysqlVersion",
		"version":     "MysqlVersion",
		"bs":          "BufferSize",
		"buffer-size": "BufferSize",
		"port":        "Port",
		"p":           "Port",
	}
}

func setInputs(inputs []string) (map[string][]string, error) {
	args := expected()
	result := make(map[string][]string)
	var err error
	var curIndex string
	for i := 0; i < len(inputs); i++ {
		v := strings.Trim(inputs[i], " ")
		if strings.Contains(v, "=") {
			err = getSplitConfigs(v, args, &result, &curIndex)
		} else {
			err = appendConfig(&curIndex, args, &result, v)
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func getSplitConfigs(v string, args map[string]string, result *map[string][]string, curIndex *string) error {
	configs := strings.Split(v, "=")
	var err error
	var errs []error
	for _, c := range configs {
		c = strings.Trim(c, " ")
		err = appendConfig(curIndex, args, result, c)
		errs = append(errs, err)
	}
	return utils.GetErrors(errs)
}

func appendConfig(curIndex *string, args map[string]string, result *map[string][]string, arg string) error {
	removeDashes(&arg)
	if args[arg] == "" {
		if (*result)[*curIndex] == nil {
			err := fmt.Errorf("the %s argument does not exist", arg)
			return err
		}
		(*result)[*curIndex] = append((*result)[*curIndex], arg)
	} else {
		*curIndex = args[arg]
		(*result)[*curIndex] = make([]string, 0)
	}
	return nil
}

func removeDashes(input *string) {
	result := strings.Replace(*input, "-", "", 2)
	*input = result
}

func (c *Configs) GetMysqlVersion() string {
	if c.inputs["MysqlVersion"] != nil && len(c.inputs["MysqlVersion"]) > 0 {
		return c.inputs["MysqlVersion"][0]
	}
	return "latest"
}

func (c *Configs) GetPort() int {
	if c.inputs["Port"] != nil && len(c.inputs["Port"]) > 0 {
		port, err := strconv.Atoi(c.inputs["Port"][0])
		if err != nil {
			return 6603
		}
		return port
	}
	return 6603
}

func (c *Configs) GetBufferSize() int {
	if c.inputs["BufferSize"] != nil && len(c.inputs["BufferSize"]) > 0 {
		size := c.inputs["BufferSize"][0]
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			return 10
		}
		return sizeInt
	}
	return 10
}
