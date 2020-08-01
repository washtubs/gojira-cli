package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"

	"github.com/pkg/errors"
)

const rpcReplyAddrEnv = "RPC_REPLY_ADDR"

type RpcClient struct {
	*rpc.Client
}

func (c *RpcClient) Query(fifo string) {
	jqlBuf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to read stdin for query"))
	}

	err = c.Call("QueryRunnerReceiver.HandleQuery", QueryReq{Jql: string(jqlBuf), File: fifo}, &Noop{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *RpcClient) LoadResults() {
	err := c.Call("QueryRunnerReceiver.LoadResults", Noop{}, &Noop{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *RpcClient) PrintIssue(issueId string) {
	resp := StringResp{}
	err := c.Call("QueryRunnerReceiver.PrintIssue", issueId, &resp)
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
