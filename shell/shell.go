package shell

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/blainemoser/TrySql/configs"
)

var Testing bool

type TrySqlShell struct {
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
	In  string
	Out string
}

func New(configs *configs.Configs) *TrySqlShell {
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
		OsInterrupt:  c,
		UserInput:    input,
		StdIn:        stdIn,
		Reader:       reader,
		WG:           &sync.WaitGroup{},
		Buffer:       list.New(),
		BufferSize:   configs.GetBufferSize(),
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

func (c *TrySqlShell) Running() {
	go c.waitForInput()
	for {
		select {
		case interruption := <-c.OsInterrupt:
			fmt.Println(" " + interruption.String())
			close(c.OsInterrupt)
			c.WG.Done()
			return
		case command := <-c.UserInput:
			c.Capture(command)
			if c.handleCommand(command) {
				return
			}
			go c.waitForInput()
		}
	}
}

func (c *TrySqlShell) handleCommand(command string) bool {
	switch command {
	case "quit", "exit", "q":
		return c.quit()
	case "help", "h":
		c.help()
	case "history":
		c.getHistory()
		return false
	case "[error]":
		return false
	default:
		c.shellOutput(command, fmt.Sprintf("> unrecognised command '%s'. Type 'help' for help", command))
	}

	return false
}

func (c *TrySqlShell) quit() bool {
	fmt.Println("> exiting...")
	close(c.UserInput)
	c.WG.Done()
	return true
}

func (c *TrySqlShell) help() {
	c.shellOutput("help", "> figure it out for yourself...")
}

func (c *TrySqlShell) shellOutput(input, msg string) {
	b := &BufferObject{
		In:  input,
		Out: msg,
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
	if err != nil {
		c.bufferError(err)
		return
	}
	userInput = strings.TrimSuffix(userInput, "\n")
	if userInput == "" {
		return
	}
	c.UserInput <- userInput
}

func (c *TrySqlShell) bufferError(err error) {
	if err.Error() == "EOF" {
		return
	}
	c.shellOutput("", "> An error occured ("+err.Error()+"), please try again")
	c.UserInput <- "[error]"
}

func (c *TrySqlShell) Push(input string) string {
	c.UserInput <- input
	result := <-c.LastCaptured
	return result
}

func (c *TrySqlShell) getHistory() {
	fmt.Println()
	message := ""
	for e := c.Buffer.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			if add, ok := e.Value.(*BufferObject); ok {
				if len(add.In) > 0 {
					message += "> " + add.In + "\n"
				}
				if len(add.Out) > 0 {
					message += ">>> " + strings.Replace(add.Out, "> ", "", 1) + "\n"
				}
			}
		}
	}
	if len(message) > 0 {
		fmt.Println(message)
	}
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

func (c *TrySqlShell) Capture(command string) {
	if len(c.LastCaptured) > 0 {
		// Drain channel
		<-c.LastCaptured
	}
	c.LastCaptured <- command
}
