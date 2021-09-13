package help

import (
	"fmt"
	"strings"
)

var info map[string]map[string]interface{} = map[string]map[string]interface{}{
	"help": {
		"alias": []string{"h"},
		"info":  "Gets information for all commands",
	},
	"history": {
		"alias": []string{"hist", "hi"},
		"info":  "Prints a history of commands and results, limited to the buffer length",
	},
	"docker-version": {
		"alias": []string{"version", "dv"},
		"info":  "Gets the Docker Version",
	},
	"quit": {
		"alias": []string{"exit"},
		"info":  "Quits the shell",
	},
	"container-details": {
		"alias": []string{"cd", "get-container-details"},
		"info":  "Gets the TrySql container's details",
	},
	"container-id": {
		"alias": []string{"cid", "get-container-id"},
		"info":  "Gets the TrySql container's ID",
	},
	"temp-password": {
		"alias": []string{"tp", "get-temporary-password"},
		"info":  "Gets the TrySql container's temporary ROOT password",
	},
	"password": {
		"alias": []string{"p", "get-password"},
		"info":  "Gets the TrySql container's ROOT password",
	},
	"query": {
		"alias": []string{"q"},
		"info":  "Executes a query against MySQL",
	},
}

var alias map[string]string = map[string]string{
	"help":                   "help",
	"h":                      "help",
	"history":                "history",
	"hi":                     "history",
	"docker-version":         "docker-version",
	"version":                "docker-version",
	"dv":                     "docker-version",
	"quit":                   "quit",
	"exit":                   "quit",
	"container-details":      "container-details",
	"cd":                     "container-details",
	"get-container-details":  "container-details",
	"container-id":           "container-id",
	"cid":                    "container-id",
	"get-container-id":       "container-id",
	"temp-password":          "temp-password",
	"tp":                     "temp-password",
	"get-temporary-password": "temp-password",
	"password":               "password",
	"p":                      "password",
	"get-password":           "password",
	"query":                  "query",
	"q":                      "query",
}

func Get(args []string) string {
	commands := getCommands(args)
	return prepCommands(commands)
}

func getCommands(args []string) map[string]map[string]interface{} {
	commands := parseCommands(args)
	result := make(map[string]map[string]interface{})
	if len(commands) < 1 {
		return info
	}
	for _, command := range commands {
		if alias[command] == "" {
			result[command] = map[string]interface{}{
				"info": fmt.Sprintf("No command '%s'", command),
			}
			continue
		}
		result[alias[command]] = info[alias[command]]
	}
	return result
}

func parseCommands(commands []string) []string {
	if len(commands) < 1 {
		return []string{}
	}
	commands = commands[1:]
	var all []string
	for i := 0; i < len(commands); i++ {
		if len(commands[i]) < 1 || commands[i] == "help" || commands[i] == "h" {
			continue
		}
		all = append(all, commands[i])
	}
	return all
}

func prepCommands(commands map[string]map[string]interface{}) string {
	result := ""
	for key, command := range commands {
		if len(result) > 0 {
			result += "\n\n"
		}
		result += formatCommand(key, command)
	}

	return result
}

func formatCommand(key string, command map[string]interface{}) string {
	if command["info"] == nil {
		return ""
	}
	result := key
	if command["alias"] != nil {
		if alias, ok := command["alias"].([]string); ok {
			if len(alias) > 0 {
				result += " | aliases: " + strings.Join(alias, ", ")
			}
		}
	}
	if info, ok := command["info"].(string); ok {
		result += "\n\t" + info
	}
	return result
}
