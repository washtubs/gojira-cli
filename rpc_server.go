package cli

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/andygrunwald/go-jira"
)

var (
	mutex             sync.Mutex
	listener          net.Listener
	socketPortDefault = 4378
	httpPath          = "fzf"

	receiver = &QueryRunnerReceiver{&FzfReceiver{}, nil, nil, nil, nil, true}
)

type QueryRunnerReceiver struct {
	*FzfReceiver
	searcher  IssueSearcher
	issueChan <-chan jira.Issue
	formatter IssueFormatter
	cancel    <-chan bool
	init      bool
}

type QueryReq struct {
	Jql, File string
}

func (q *QueryRunnerReceiver) HandleQuery(req QueryReq, resp *Noop) error {
	jql, file := req.Jql, req.File
	log.Printf("REQUEST RECEIVED %s %s", jql, file)
	if !q.init {
		// If we already initialized we need to search again
		q.searcher.SetSearchQuery(jql)
		q.issueChan, q.interactor = q.searcher.SearchAsync()
	} // if this is the first run, the query has already been set and search has started
	q.init = false

	log.Printf("Opening fifo %s", file)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Failed to open file for append: %s", file)
		return err
	}

	formatters := mapIssueChan(q.issueChan, q.formatter)
	log.Printf("Writing")
	fzfWrite(formatters, f, q.cancel)

	// We are not concerned with reading the output of fzf here, just writing results,
	// as the user may discard them

	return nil
}

func SetupRpc() {
	srv := rpc.NewServer()
	srv.Register(receiver)
	srv.HandleHTTP(httpPath, httpPath+"debug")
}

func ListenRpcQueryRunner(searcher IssueSearcher, issueChan <-chan jira.Issue, interactor SearchInteractor, formatter IssueFormatter, cancel <-chan bool) (int, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// TODO: use a different port if needed
	port := socketPortDefault

	stopListenRpcIfNeeded(port)

	receiver.interactor = interactor
	receiver.issueChan = issueChan
	receiver.searcher = searcher
	receiver.formatter = formatter
	receiver.init = true

	//os.Setenv(socketAddrEnv, "localhost:"+strconv.Itoa(socketPortDefault))
	var err error
	listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		return 0, err
	}
	go http.Serve(listener, nil)

	return port, nil

}

func ListenRpc(f SearchInteractor) (int, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// TODO: use a different port if needed
	port := socketPortDefault

	stopListenRpcIfNeeded(port)

	receiver.interactor = f

	//os.Setenv(socketAddrEnv, "localhost:"+strconv.Itoa(socketPortDefault))
	var err error
	listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		return 0, err
	}
	go http.Serve(listener, nil)

	return port, nil
}

func stopListenRpcIfNeeded(port int) {
	if listener != nil {
		listener.Close()
	}
	listener = nil

	receiver.interactor = nil
	receiver.init = true
	receiver.searcher = nil
	receiver.issueChan = nil
	receiver.formatter = nil
	receiver.cancel = nil
}

func StopListenRpc(port int) {
	mutex.Lock()
	defer mutex.Unlock()

	stopListenRpcIfNeeded(port)

}

func SetupEnvForRpc(cmd *exec.Cmd, port int) {
	cmd.Env = append(cmd.Env, rpcReplyAddrEnv+"=localhost:"+strconv.Itoa(port))
}
