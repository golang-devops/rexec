package comms

import (
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
	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Out, reply.Error = cmd.CombinedOutput()
	return nil
}

func (e *Executor) Start(executeArgs *ExecutorExecuteArgs, reply *ExecutorStartReply) error {
	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Error = cmd.Start()
	reply.Pid = cmd.Process.Pid
	return nil
}

func init() {
	rpc.Register(new(Executor))
}
