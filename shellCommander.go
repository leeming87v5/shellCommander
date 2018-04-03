/*
Package shellCommander provides a handy way to organize the shell commands to run.
We provide an easy way to mimic the conventional unix-style shell facility -- pipe.
Example:
    // First, specify the commands to be piped up.
    c1 := NewCommand(NoCleanFn, "ps", "-ef")
    c2 := NewCommand(NoCleanFn, "grep", "lantern")
    c3 := NewCommand(NoCleanFn, "grep", "-v", "grep")

    // Then, create an ordered command sequence with NewPipeCmd().
    pc := NewPipeCmd(c1, c2, c3)

    // Create a context (used to set timeout etc.), and run the piped commands.
    ctx := context.Background()
    stdout, stderr, err := pc.Run(ctx)

    // Optionally, clean the intermediate result
    errc := pc.Clean()
*/
package shellCommander

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"golang.org/x/net/context"
)

type Command interface {
	Name() string
	Params() []string
	Clean() error
}

type Cleaner func(Command) error

var (
	NoCleanFn = func(c Command) error { return nil }
	debug     = true
)

type cmd struct {
	name    string
	params  []string
	cleanfn Cleaner
}

func (this *cmd) Name() string {
	return this.name
}

func (this *cmd) Params() []string {
	return this.params
}

func (this *cmd) Clean() error {
	return this.cleanfn(this)
}

func NewCommand(cleanfn Cleaner, name string, args ...string) Command {
	c := &cmd{
		name:    name,
		params:  args,
		cleanfn: cleanfn,
	}
	return c
}

type PipeCmd interface {
	Run(ctx context.Context) (stdout string, stderr string, err error)
	Clean() error
}

func NewPipeCmd(cmds ...Command) PipeCmd {
	pc := &pipeCmd{
		stdin:   &bytes.Buffer{},
		stdout:  &bytes.Buffer{},
		stderr:  &bytes.Buffer{},
		subCmds: make([]Command, 0),
	}
	for _, c := range cmds {
		pc.subCmds = append(pc.subCmds, c)
	}
	return pc
}

type pipeCmd struct {
	stdin   *bytes.Buffer
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	subCmds []Command
}

func (this *pipeCmd) Run(ctx context.Context) (stdout string, stderr string, err error) {
	// for k, c := range this.subCmds {
	// 	cmd := exec.Command(c.Name(), c.Params()...)
	// 	// show the command to run
	// 	fmt.Println(cmd.Args)

	// 	stdin, erri := cmd.StdinPipe()
	// 	if erri != nil {
	// 		return "", "", erri
	// 	}
	// 	stdout, erro := cmd.StdoutPipe()
	// 	if erro != nil {
	// 		return "", "", erro
	// 	}
	// }
	for _, c := range this.subCmds {
		this.stdout = &bytes.Buffer{}
		this.stderr = &bytes.Buffer{}
		cmd := exec.Command(c.Name(), c.Params()...)
		cmd.Stdin = this.stdin
		cmd.Stdout = this.stdout
		cmd.Stderr = this.stderr

		// show the command to run
		if debug {
			fmt.Println(cmd.Args)
		}
		err = cmd.Start()
		if err != nil {
			return "", "", err
		}
		tchan := make(chan int, 1)
		go func() {
			err = cmd.Wait()
			if err != nil {
				tchan <- 0
			}
			tchan <- 1
		}()

		select {
		case i := <-tchan:
			if i == 0 {
				return "", "", err
			}
		case <-ctx.Done():
			err := cmd.Process.Kill()
			if err != nil {
				return "", "", err
			}
			return "", "", ctx.Err()
		}
		// fmt.Println(this.stdout.String())
		this.stdin = cmd.Stdout.(*bytes.Buffer)
	}
	return this.stdout.String(), this.stderr.String(), nil
}

func (this *pipeCmd) Clean() error {
	var errc error
	for _, c := range this.subCmds {
		if err := c.Clean(); err != nil {
			if errc != nil {
				errc = errors.New(errc.Error() + "\n" + err.Error())
			} else {
				errc = errors.New(err.Error())
			}
		}
	}
	return errc
}
