package comms

import (
	"fmt"
	"net/rpc"
)

type Client struct {
	rpcClient *rpc.Client
}

func NewConnectedClient(serverAddress string) (*Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial tcp server (%s), error: %s", serverAddress, err.Error())
	}

	return &Client{
		rpcClient: client,
	}, nil
}

func (c *Client) Execute(args *ExecutorExecuteArgs, reply *ExecutorExecuteReply) error {
	return c.rpcClient.Call("Executor.Execute", args, reply)
}

func (c *Client) Start(args *ExecutorExecuteArgs, reply *ExecutorStartReply) error {
	return c.rpcClient.Call("Executor.Start", args, reply)
}
