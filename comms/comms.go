package comms

//Comms is the shared interface between Client and Executor/server
type Comms interface {
	Execute(executeArgs *ExecutorExecuteArgs, reply *ExecutorExecuteReply) error
	Start(executeArgs *ExecutorExecuteArgs, reply *ExecutorStartReply) error
	GetFeedback(args *GetFeedbackArgs, reply *GetFeedbackReply) error
}

type ExecutorExecuteArgs struct {
	Exe  string
	Args []string
}

type ExecutorExecuteReply struct {
	Out   []byte
	Error error
}

type ExecutorStartReply struct {
	Pid       int
	SessionID string
	Error     error
}

type GetFeedbackArgs struct {
	SessionID   string
	OffsetLines int
}

type GetFeedbackReply struct {
	Lines           []string
	NextOffsetLines int
	Error           error
}
