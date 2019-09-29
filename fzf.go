package cli

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
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

type CandidatesGenerater interface {
	Next() (Formatter, bool)
}

type SliceCandidatesGenerater struct {
	candidates []Formatter
	ptr        int
}

func (s *SliceCandidatesGenerater) Next() (Formatter, bool) {
	if s.ptr >= len(s.candidates) {
		return nil, false
	}
	f := s.candidates[s.ptr]
	s.ptr = s.ptr + 1
	return f, true
}

type ChannelCandidatesGenerator struct {
	channel <-chan Formatter
}

func (c *ChannelCandidatesGenerator) Next() (Formatter, bool) {
	f, more := <-c.channel
	return f, more
}

func FzfSelect(candidates []Formatter, opts SelectOptions) ([]int, bool, error) {
	return FzfSelectWithGenerator(&SliceCandidatesGenerater{candidates, 0}, opts)
}

func FzfSelectChan(c <-chan Formatter, opts SelectOptions) ([]int, bool, error) {
	return FzfSelectWithGenerator(&ChannelCandidatesGenerator{c}, opts)
}

func FzfSelectWithGenerator(candidates CandidatesGenerater, opts SelectOptions) ([]int, bool, error) {
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

	// Collect all output in a buffer
	var stdout *bytes.Buffer
	//cmd.Stdout = &stdout

	buf := bytes.NewBuffer([]byte{})
	inR, inW := io.Pipe()
	cmd.Stdin = inR

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, false, errors.Wrap(err, "error starting command")
	}

	go func() {
		line := 0
		for {
			formatter, found := candidates.Next()
			if !found {
				inW.Close()
				break
			}
			buf.WriteString(strconv.Itoa(line) + " " + formatter.Format() + "\n")
			buf.WriteTo(inW)
			line = line + 1
		}
	}()

	go func() {
		out, err := ioutil.ReadAll(outPipe)
		if err != nil {
			log.Println(err)
			return
		}

		stdout = bytes.NewBuffer(out)
		log.Println("output complete")
	}()
	err = cmd.Wait()
	log.Println("command complete")
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
