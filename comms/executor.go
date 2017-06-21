package comms

import (
	"fmt"
	"net/rpc"
	"os/exec"

	"github.com/golang-devops/rexec/logging"
)

//Executor is the RPC server/executor
type Executor struct {
	sessions *executorSessions
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

	session := e.sessions.NewSession()
	reply.SessionID = session.id
	logger = logger.WithField("session-id", session.id)
	logger.Debug(fmt.Sprintf("Created session %s, currently %d sessions running", session.id, e.sessions.SessionCount()))

	cmd := exec.Command(executeArgs.Exe, executeArgs.Args...)
	session.SetCommand(cmd)

	if err := session.StartCommand(logger); err != nil {
		logger.WithError(err).Error("Error starting command")
		reply.Error = err
		return err
	}

	reply.Pid = cmd.Process.Pid
	pidStr = fmt.Sprintf("%d", reply.Pid)

	return nil
}

func (e *Executor) GetFeedback(args *GetFeedbackArgs, reply *GetFeedbackReply) error {
	logger := logging.Logger()

	session, err := e.sessions.GetSession(args.SessionID)
	if err != nil {
		logger.WithError(err).Error("Error getting session")
		reply.Error = err
		return err
	}

	offsetLines := args.OffsetLines
	lines, err := session.ReadNextLines(offsetLines)
	if err != nil {
		if !IsEOFAndExitedErr(err) {
			logger.WithError(err).Error("Error reading lines")
		}
		reply.Error = err
		return err
	}

	reply.Lines = lines
	reply.NextOffsetLines = offsetLines + len(lines)
	return nil
}

var _ Comms = &Executor{} //type-safety

func init() {
	rpc.Register(&Executor{
		sessions: NewExecutorSessions(),
	})
}
