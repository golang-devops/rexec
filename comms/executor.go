package comms

import (
	"fmt"
	"net/rpc"
	"os/exec"
)

type Executor int

type ExecutorExecuteArgs struct {
	Exe  string
	Args []string
}

type ExecutorExecuteReply struct {
	Out   []byte
	Error error
}

type ExecutorStartReply struct {
	Pid   int
	Error error
}

func (e *Executor) Execute(executeArgs *ExecutorExecuteArgs, reply *ExecutorExecuteReply) error {
	fmt.Println(fmt.Sprintf("Execute. %q", append([]string{executeArgs.Exe}, executeArgs.Args...)))

	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Out, reply.Error = cmd.CombinedOutput()
	return nil
}

func (e *Executor) Start(executeArgs *ExecutorExecuteArgs, reply *ExecutorStartReply) error {
	pidStr := "NOT_SET"
	defer func() {
		fmt.Println(fmt.Sprintf("Start (pid = %s). %q", pidStr, append([]string{executeArgs.Exe}, executeArgs.Args...)))
	}()

	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Error = cmd.Start()
	reply.Pid = cmd.Process.Pid

	pidStr = fmt.Sprintf("%d", reply.Pid)

	return nil
}

func init() {
	rpc.Register(new(Executor))
}
