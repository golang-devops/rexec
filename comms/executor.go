package comms

import (
	"fmt"
	"net/rpc"
	"os/exec"

	"github.com/golang-devops/rexec/logging"
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
	logger := logging.Logger()

	logger.Info(fmt.Sprintf("Execute. %q", append([]string{executeArgs.Exe}, executeArgs.Args...)))

	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Out, reply.Error = cmd.CombinedOutput()
	if reply.Error != nil {
		out := ""
		if reply.Out != nil {
			out = " Output was: " + string(reply.Out)
		}
		return fmt.Errorf("Failed to Execute, error: %s.%s", reply.Error.Error(), out)
	}
	return nil
}

func (e *Executor) Start(executeArgs *ExecutorExecuteArgs, reply *ExecutorStartReply) error {
	logger := logging.Logger()

	pidStr := "NOT_SET"
	defer func() {
		logger.Info(fmt.Sprintf("Start (pid = %s). %q", pidStr, append([]string{executeArgs.Exe}, executeArgs.Args...)))
	}()

	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	reply.Error = cmd.Start()
	if reply.Error != nil {
		return reply.Error
	}
	reply.Pid = cmd.Process.Pid
	pidStr = fmt.Sprintf("%d", reply.Pid)

	return nil
}

func init() {
	rpc.Register(new(Executor))
}
