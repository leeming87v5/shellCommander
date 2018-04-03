package shellCommander

import (
	"errors"
	"fmt"
	"testing"

	"golang.org/x/net/context"
)

func Test_SingleCmdPipe(t *testing.T) {
	c1 := NewCommand(NoCleanFn, "/usr/local/hadoop/hadoop-2.9.0/bin/hdfs", "dfs", "-help")

	pc := NewPipeCmd(c1)

	ctx := context.Background()
	stdout, stderr, err := pc.Run(ctx)
	errc := pc.Clean()
	if err != nil || errc != nil {
		fmt.Println("Fail")
		fmt.Println(err)
		fmt.Println(errc)
		fmt.Println(stderr)
	} else {
		fmt.Println("Success")
		fmt.Println(stdout)
	}

}

func Test_PipeRun(t *testing.T) {
	c1 := NewCommand(NoCleanFn, "ps", "-ef")
	c2 := NewCommand(NoCleanFn, "grep", "watchdog")
	c3 := NewCommand(NoCleanFn, "grep", "-v", "grep")

	pc := NewPipeCmd(c1, c2, c3)

	ctx := context.Background()
	stdout, stderr, err := pc.Run(ctx)
	errc := pc.Clean()
	if err != nil || errc != nil {
		fmt.Println("Fail")
		fmt.Println(err)
		fmt.Println(errc)
		fmt.Println(stderr)
	} else {
		fmt.Println("Success")
		fmt.Println(stdout)
	}
}

func Test_PipeRunCleanErr(t *testing.T) {
	cerrfn := func(c Command) error {
		errstr := c.Name()
		for _, p := range c.Params() {
			errstr += " "
			errstr += p
		}
		return errors.New(errstr)
	}
	c1 := NewCommand(cerrfn, "ps", "-ef")
	c2 := NewCommand(cerrfn, "grep", "watchdog")
	c3 := NewCommand(cerrfn, "grep", "-v", "grep")

	pc := NewPipeCmd(c1, c2, c3)

	ctx := context.Background()
	stdout, stderr, err := pc.Run(ctx)
	errc := pc.Clean()
	if err != nil || errc != nil {
		fmt.Println("Fail")
		fmt.Println(err)
		fmt.Println(errc)
		fmt.Println(stderr)
	} else {
		fmt.Println("Success")
		fmt.Println(stdout)
	}
}
