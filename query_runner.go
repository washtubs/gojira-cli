package cli

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

func createQueryFile(initialJql string) string {
	f, err := ioutil.TempFile("", "gojira-cli")
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to create query file"))
	}

	defer f.Close()
	// TODO: does this write all of it?
	_, err = f.WriteString(initialJql)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to write JQL to query file"))
	}

	return f.Name()
}

func executeQueryRunner(searcher IssueSearcher, formatter IssueFormatter, issuesChan <-chan jira.Issue, interactor SearchInteractor, opts SelectOptions) ([]int, bool, error) {
	log.Printf("STARTING QUERY RUNNER")

	cancel := make(chan bool, 1)
	port, err := ListenRpcQueryRunner(searcher, issuesChan, interactor, formatter, cancel)
	if err != nil {
		return nil, false, errors.Wrap(err, "Failed to start listening to RPC")
	}
	defer StopListenRpc(port)

	args := []string{"-d", "70%", "--"}
	args = appendFzfArgs(args, opts, port)
	cmd := exec.Command("query-runner", args...)

	queryFile := createQueryFile(searcher.GetSearchQuery())

	cmd.Env = os.Environ()
	SetupEnvForRpc(cmd, port)

	cmd.Env = append(cmd.Env, "QUERY_RUNNER_FILE="+queryFile)
	cmd.Env = append(cmd.Env, "QUERY_RUNNER_COMMAND=gojira-cli _rpc query")
	cmd.Env = append(cmd.Env, "QUERY_RUNNER_USE_COMMAND_FIFO=1")

	// Collect all output in a buffer
	var stdout *bytes.Buffer
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, false, errors.Wrap(err, "error starting command")
	}

	go func() {

		// When Fzf closes stdout,
		// we need to cancel the formatters stream,
		// and close stdin,
		// so Wait() will finish
		out, err := ioutil.ReadAll(outPipe)
		if err != nil {
			log.Println(err)
			return
		}

		stdout = bytes.NewBuffer(out)
		cancel <- true
	}()

	err = cmd.Wait()

	if err != nil {
		return nil, true, nil
	}

	indexes := fzfConvertOutput(stdout.String())
	log.Printf("query-runner finished successfully: %d records", len(indexes))
	return indexes, false, nil
}
