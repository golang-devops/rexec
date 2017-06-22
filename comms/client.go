package comms

import (
	"fmt"
	"net/rpc"

	"github.com/pkg/errors"
)

func NewConnectedClient(serverAddress string) (*Client, error) {
	rpcClient, err := rpc.DialHTTP("tcp", serverAddress)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial tcp server (%s), error: %s", serverAddress, err.Error())
	}

	return &Client{
		rpcClient: rpcClient,
	}, nil
}

type Client struct {
	rpcClient *rpc.Client
}

func (c *Client) Execute(args *ExecutorExecuteArgs, reply *ExecutorExecuteReply) error {
	return c.rpcClient.Call("Executor.Execute", args, reply)
}

func (c *Client) Start(args *ExecutorExecuteArgs, reply *ExecutorStartReply) error {
	return c.rpcClient.Call("Executor.Start", args, reply)
}

func (c *Client) GetFeedback(args *GetFeedbackArgs, reply *GetFeedbackReply) error {
	return c.rpcClient.Call("Executor.GetFeedback", args, reply)
}

func (c *Client) RunWithFeedback(args *ExecutorExecuteArgs, onFeedback func(lines []string)) error {
	var reply ExecutorStartReply
	if err := c.Start(args, &reply); err != nil {
		return errors.Wrap(err, "Failed to start")
	} else if reply.Error != nil {
		return errors.Wrap(reply.Error, "Failed to start command (but got reply)")
	}

	feedbackArgs := &GetFeedbackArgs{
		SessionID:   reply.SessionID,
		OffsetLines: 0,
	}
	for {
		var feedbackReply GetFeedbackReply
		if err := c.GetFeedback(feedbackArgs, &feedbackReply); err != nil {
			if IsEOFAndExitedSuccessfullyErr(err) {
				break
			}
			return errors.Wrap(err, "Failed to get feedback")
		} else if feedbackReply.Error != nil {
			return errors.Wrap(feedbackReply.Error, "Failed to get feedback (but got reply)")
		}

		if len(feedbackReply.Lines) == 0 {
			continue
		}

		onFeedback(feedbackReply.Lines)
		feedbackArgs.OffsetLines = feedbackReply.NextOffsetLines
	}

	return nil
}

var _ Comms = &Client{} //type-safety
