package cli

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

var (
	mutex             sync.Mutex
	listener          net.Listener
	socketPortDefault = 4378
	httpPath          = "fzf"
	fzfReceiver       = &FzfReceiver{}
)

type FzfReceiver struct {
	h SearchInteractor
}

type Noop struct{}

type StringResp struct {
	String string
}

func (f *FzfReceiver) LoadResults(req Noop, resp *Noop) error {
	if f.h == nil {
		log.Println("Fzf request received but not registered")
		return errors.New("Not registered")
	}
	log.Println("LoadResults")
	f.h.LoadResults()
	return nil
}

func (f *FzfReceiver) PrintIssue(issueId string, resp *StringResp) error {
	if f.h == nil {
		log.Println("Fzf request received but not registered")
		return errors.New("Not registered")
	}

	found := false
	for _, issue := range f.h.Loaded() {
		if issue.ID == issueId {
			resp.String = PrintIssue(issue)
			found = true
			break
		}
	}

	if !found {
		return errors.New("No issue found")
	}
	return nil
}

func SetupRpc() {
	srv := rpc.NewServer()
	srv.Register(fzfReceiver)
	srv.HandleHTTP(httpPath, httpPath+"debug")
}

func ListenRpc(f SearchInteractor) (int, error) {
	log.Println("acquiring mutex")
	mutex.Lock()
	defer mutex.Unlock()

	// TODO: use a different port if needed
	port := socketPortDefault

	stopListenRpcIfNeeded(port)

	fzfReceiver.h = f

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

	fzfReceiver.h = nil
}

func StopListenRpc(port int) {
	log.Println("acquiring mutex")
	mutex.Lock()
	defer mutex.Unlock()

	stopListenRpcIfNeeded(port)

}
