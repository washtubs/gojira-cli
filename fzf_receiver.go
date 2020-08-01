package cli

import (
	"log"

	"github.com/pkg/errors"
)

type FzfReceiver struct {
	interactor SearchInteractor
}

type Noop struct{}

type StringResp struct {
	String string
}

func (f *FzfReceiver) LoadResults(req Noop, resp *Noop) error {
	if f.interactor == nil {
		log.Println("Fzf request received but not registered")
		return errors.New("Not registered")
	}
	log.Println("LoadResults")
	f.interactor.LoadResults()
	return nil
}

func (f *FzfReceiver) PrintIssue(issueId string, resp *StringResp) error {
	if f.interactor == nil {
		log.Println("Fzf request received but not registered")
		return errors.New("Not registered")
	}

	found := false
	for _, issue := range f.interactor.Loaded() {
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
