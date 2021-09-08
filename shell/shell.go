package shell

import (
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/blainemoser/TrySql/help"
	"github.com/blainemoser/TrySql/trysql"
	"github.com/blainemoser/TrySql/utils"
)

var Testing bool

const timeFormat string = "15:04:05"
const shellVersion string = "1.0.0"

type TrySqlShell struct {
	TS           *trysql.TrySql
	OsInterrupt  chan os.Signal
	UserInput    chan string
	LastCaptured chan string
	StdIn        io.Reader
	Reader       *bufio.Reader
	WG           *sync.WaitGroup
	Buffer       *list.List
	BufferSize   int
}

type BufferObject struct {
	In   string
	Out  string
	Time time.Time
}

func New(ts *trysql.TrySql) *TrySqlShell {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	input := make(chan string)
	var stdIn io.Reader
	if Testing {
		var b []byte
		stdIn = bytes.NewReader(b)
	} else {
		stdIn = os.Stdin
	}
	reader := bufio.NewReader(stdIn)
	return &TrySqlShell{
		TS:           ts,
		OsInterrupt:  c,
		UserInput:    input,
		StdIn:        stdIn,
		Reader:       reader,
		WG:           &sync.WaitGroup{},
		Buffer:       list.New(),
		BufferSize:   ts.Configs.GetBufferSize(),
		LastCaptured: make(chan string, 1),
	}
}

func (c *TrySqlShell) Start() {
	c.WG.Add(1)
	go c.Running()
	c.WG.Wait()
}

func (c *TrySqlShell) StartTest() {
	go c.Running()
}

func (c *TrySqlShell) greeting() {
	msg := fmt.Sprintf(" TrySql Shell version %s ", shellVersion)
	var line string
	for i := 0; i < len(msg); i++ {
		line += "_"
	}
	fmt.Print("\t" + line + "\n\n\t" + msg + "\n\t" + line + "\n\n")
}

func (c *TrySqlShell) Running() {
	c.greeting()
	go c.waitForInput()
	for {
		select {
		case interruption := <-c.OsInterrupt:
			fmt.Println(" " + interruption.String())
			close(c.OsInterrupt)
			c.WG.Done()
			return
		case command := <-c.UserInput:
			c.Capture(&command)
			if c.handleCommand(command) {
				return
			}
			go c.waitForInput()
		}
	}
}

func (c *TrySqlShell) handleCommand(command string) bool {
	if commandSplit, ok := c.checkHelp(command); ok {
		c.help(commandSplit)
		return false
	}
	switch command {
	case "":
		return false
	case "quit", "exit", "q":
		return c.quit()
	case "history", "hist", "hi":
		return c.getHistory()
	case "docker-version", "version", "dv":
		return c.getVersion(command)
	case "[error]":
		return false
	default:
		c.shellOutput(command, fmt.Sprintf("> unrecognised command '%s'. Type 'help' for help", strings.ReplaceAll(command, "\n", "")))
	}

	return false
}

func (c *TrySqlShell) quit() bool {
	fmt.Println("> exiting...")
	close(c.UserInput)
	c.WG.Done()
	return true
}

func (c *TrySqlShell) help(command []string) {
	result := help.Get(command)
	c.shellOutput("", "\n"+result+"\n")
}

func (c *TrySqlShell) checkHelp(command string) ([]string, bool) {
	split := strings.Split(command, " ")
	if len(split) < 1 {
		return []string{}, false
	}
	if split[0] == "help" || split[0] == "h" {
		return split, true
	}
	return []string{}, false
}

func (c *TrySqlShell) shellOutput(input, msg string) {
	if len(input) < 1 {
		fmt.Println(msg)
		return
	}
	b := &BufferObject{
		In:   input,
		Out:  msg,
		Time: time.Now(),
	}
	if c.Buffer.Len() >= c.BufferSize {
		e := c.Buffer.Back()
		c.Buffer.Remove(e)
	}
	c.Buffer.PushFront(b)
	fmt.Println(msg)
}

func (c *TrySqlShell) waitForInput() {
	fmt.Print("> ")
	userInput, err := c.Reader.ReadString('\n')
	if err != nil && userInput == "\n" {
		err = errors.New("carriage_return")
	}
	if err != nil {
		c.bufferError(err)
		return
	}
	c.special(&userInput)
	if userInput == "" {
		c.UserInput <- "[error]"
	}
	userInput = strings.TrimSuffix(userInput, "\n")
	c.UserInput <- userInput
}

func (c *TrySqlShell) bufferError(err error) {
	switch err.Error() {
	case "EOF":
		return
	case "carriage_return":
		c.UserInput <- "[error]"
		return
	default:
		c.shellOutput("", "> An error occured ("+err.Error()+"), please try again")
		c.UserInput <- "[error]"
	}
}

func (c *TrySqlShell) Push(input string) string {
	c.UserInput <- input
	result := <-c.LastCaptured
	return result
}

func (c *TrySqlShell) getHistory() bool {
	fmt.Println()
	message := ""
	for e := c.Buffer.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			if add, ok := e.Value.(*BufferObject); ok {
				if len(add.In) > 0 {
					message += "\t" + add.In + "\n"
				}
				if len(add.Out) > 0 {
					outMsg := strings.Replace(add.Out, "> ", "", 1)
					utils.TruncString(&outMsg, 100)
					message += "\t-> " + outMsg + "\n\t   at " + add.Time.Format(timeFormat) + "\n\n"
				}
			}
		}
	}
	if len(message) > 0 {
		fmt.Println(message)
	}
	return false
}

func (c *TrySqlShell) LastOutput() string {
	e := c.Buffer.Front()
	if e == nil {
		return ""
	}
	if e.Value == nil {
		return ""
	}
	if lastBuffer, ok := e.Value.(*BufferObject); ok {
		return lastBuffer.Out
	}
	return ""
}

func (c *TrySqlShell) getVersion(command string) bool {
	c.shellOutput(command, "> "+c.TS.DockerVersion())
	return false
}

func (c *TrySqlShell) Capture(command *string) {
	c.sanitize(command)
	if len(c.LastCaptured) > 0 {
		// Drain channel
		<-c.LastCaptured
	}
	c.LastCaptured <- *command
}

func (c *TrySqlShell) lastCommand() string {
	if c.Buffer.Len() < 1 {
		return ""
	}
	e := c.Buffer.Front()
	if e == nil {
		return ""
	}
	if obj, ok := e.Value.(*BufferObject); ok {
		return obj.In
	}
	return ""
}

func (c *TrySqlShell) sanitize(command *string) {
	*command = strings.ReplaceAll(*command, "\n", "")
	*command = strings.ReplaceAll(*command, "\t", "")
	*command = strings.Trim(*command, " ")
}

func (c *TrySqlShell) special(userInput *string) {
	bytes := []byte(*userInput)
	if len(bytes) >= 3 {
		if bytes[0] == 27 && bytes[1] == 91 {
			if bytes[2] == 65 {
				lst := c.lastCommand()
				*userInput = lst
				return
			}
			*userInput = ""
		}
	}
}
