package cli

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

const rpcReplyAddrEnv = "RPC_REPLY_ADDR"

type RpcClient struct {
	*rpc.Client
}

func (c *RpcClient) LoadResults() {
	err := c.Call("FzfReceiver.LoadResults", Noop{}, &Noop{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *RpcClient) PrintIssue(issueId string) {
	resp := StringResp{}
	err := c.Call("FzfReceiver.PrintIssue", issueId, &resp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(resp.String)
}

func NewRpcClient() *RpcClient {
	socketAddr := os.Getenv(rpcReplyAddrEnv)
	if socketAddr == "" {
		socketAddr = fmt.Sprintf("localhost:%d", socketPortDefault)
	}
	client, err := rpc.DialHTTPPath("tcp", socketAddr, httpPath)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return &RpcClient{client}
}
