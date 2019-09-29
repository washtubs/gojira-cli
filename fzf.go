package cli

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// responsible for selecting
// actionBases
// issues

type SelectOptions struct {
	Prompt string
	One    bool
}

func formatterSliceToChan(formatters []Formatter) chan Formatter {
	channel := make(chan Formatter, len(formatters))
	for _, f := range formatters {
		channel <- f
	}
	close(channel)
	return channel
}

func FzfSelect(candidates []Formatter, opts SelectOptions, rpcPort int) ([]int, bool, error) {
	return FzfSelectChan(formatterSliceToChan(candidates), opts, rpcPort)
}

func FzfSelectChan(c <-chan Formatter, opts SelectOptions, rpcPort int) ([]int, bool, error) {
	args := []string{
		"--with-nth", "2..", "--reverse",
	}
	if opts.One {
		args = append(args, "+m")
	}
	if opts.Prompt != "" {
		args = append(args, "--prompt="+opts.Prompt)
	}

	if rpcPort != 0 { // RPC supported
		args = append(args, "--preview", "gojira-cli _rpc print {}")
		args = append(args, "--bind", "f1:execute(gojira-cli _rpc load)")
	}

	cmd := exec.Command("fzf-tmux", args...) // TODO figure out why I have to use fzf-tmux instead of fzf

	if rpcPort != 0 { // RPC supported
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, rpcReplyAddrEnv+"=localhost:"+strconv.Itoa(rpcPort))
	}

	// Collect all output in a buffer
	var stdout *bytes.Buffer
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer([]byte{})
	inR, inW := io.Pipe()
	cmd.Stdin = inR

	err = cmd.Start()
	if err != nil {
		return nil, false, errors.Wrap(err, "error starting command")
	}

	cancel := make(chan bool, 1)
	go func() {
		line := 0
		for {
			select {
			case formatter, more := <-c:
				if !more {
					log.Println("closed formatters")
					inW.Close()
					return
				}
				buf.WriteString(strconv.Itoa(line) + " " + formatter.Format() + "\n")
				buf.WriteTo(inW)
				line = line + 1
			case <-cancel:
				inW.Close()
				return
			}
		}
	}()

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

	// convert full string results into indexes
	results := strings.Split(stdout.String(), "\n")
	indexes := make([]int, 0, len(results))
	for _, result := range results {
		if result == "" {
			continue
		}
		idx, err := strconv.Atoi(strings.Fields(result)[0])
		if err != nil {
			panic(err)
		}

		indexes = append(indexes, idx)
	}

	return indexes, false, nil
}
