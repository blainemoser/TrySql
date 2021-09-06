package configs

import (
	"fmt"
	"strconv"
	"strings"
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
		"v":             "MysqlVersion",
		"--version":     "MysqlVersion",
		"bs":            "BufferSize",
		"--buffer-size": "BufferSize",
	}
}

func setInputs(inputs []string) (map[string][]string, error) {
	args := expected()
	result := map[string][]string{}
	var err error
	var curIndex string
	for i := 0; i < len(inputs); i++ {
		v := strings.Trim(inputs[i], " ")
		if strings.Contains(v, "=") {
			configs := strings.Split(v, "=")
			err = appendConfig(&curIndex, args, &result, configs[1], configs[0])
		} else {
			err = appendConfig(&curIndex, args, &result, v, v)
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func appendConfig(curIndex *string, args map[string]string, result *map[string][]string, value, index string) error {
	removeDashes(&index)
	if args[index] == "" {
		if (*result)[*curIndex] == nil {
			err := fmt.Errorf("the %s argument does not exist", index)
			return err
		}
		(*result)[*curIndex] = append((*result)[*curIndex], value)
	} else {
		*curIndex = args[index]
		(*result)[*curIndex] = make([]string, 0)
	}
	return nil
}

func removeDashes(input *string) {
	result := strings.ReplaceAll(*input, "-", "")
	*input = result
}

func (c *Configs) GetMysqlVersion() string {
	if c.inputs["MysqlVersion"] != nil && len(c.inputs["MysqlVersion"]) > 0 {
		return c.inputs["MysqlVersion"][0]
	}
	return "latest"
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
