package cli

import (
	"bytes"
	"io"
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
		"--with-nth", "2..",
	}
	if opts.One {
		args = append(args, "+m")
	}
	if opts.Prompt != "" {
		args = append(args, "--prompt="+opts.Prompt)
	}

	cmd := exec.Command("fzf-tmux", args...) // TODO figure out why I have to use fzf-tmux instead of fzf

	if rpcPort != 0 {
		// RPC supported
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, rpcReplyAddrEnv+"=localhost:"+strconv.Itoa(rpcPort))
		// TODO setup actions / keybindings
	}

	// Collect all output in a buffer
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	buf := bytes.NewBuffer([]byte{})
	inR, inW := io.Pipe()
	cmd.Stdin = inR

	err := cmd.Start()
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

	err = cmd.Wait()
	log.Println("cancelling")
	cancel <- true

	log.Println("a")
	if err != nil {
		return nil, true, nil
	}

	log.Println("b")
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

	log.Println("c")
	return indexes, false, nil
}
