package cli

import (
	"log"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

type IssueSelector struct {
	formatter IssueFormatter
}

func FilterIssuesByIds(issues []jira.Issue, idxs []int) []jira.Issue {
	filtered := make([]jira.Issue, 0, len(idxs))
	for _, idx := range idxs {
		filtered = append(filtered, issues[idx])
	}
	return filtered
}

func mapIssueChan(issues <-chan jira.Issue, formatter IssueFormatter) <-chan Formatter {
	formatters := make(chan Formatter, len(issues))
	go func() {
		for {
			issue, more := <-issues
			if !more {
				log.Printf("closed")
				close(formatters)
				return
			}
			log.Printf("got formatter")
			formatters <- IssueFormattable{issue, formatter}
			log.Printf("put formatter")
		}
	}()
	return formatters
}

type fixedSearchInteractor struct {
	issues []jira.Issue
}

func (f *fixedSearchInteractor) Loaded() []jira.Issue    { return f.issues }
func (f *fixedSearchInteractor) Append(issue jira.Issue) { panic("not implemented") }
func (f *fixedSearchInteractor) LoadResults()            {}
func (f *fixedSearchInteractor) CloseSearch()            {}

func (s *IssueSelector) SelectSlc(issues []jira.Issue, opts SelectOptions) ([]int, bool, error) {
	issueChan := make(chan jira.Issue, len(issues))
	for _, issue := range issues {
		issueChan <- issue
	}
	close(issueChan)
	return s.Select(issueChan, &fixedSearchInteractor{issues}, opts)
}

func (s *IssueSelector) Select(issues <-chan jira.Issue, interactor SearchInteractor, opts SelectOptions) ([]int, bool, error) {
	port, err := ListenRpc(interactor)
	if err != nil {
		return nil, false, errors.Wrap(err, "Failed to start listening to RPC")
	}
	defer StopListenRpc(port)

	idxs, canceled, err := FzfSelectChan(mapIssueChan(issues, s.formatter), opts, port)
	log.Println("done select")
	if err != nil {
		return nil, false, errors.Wrap(err, "Failed to get FZF results for issue selection")
	}
	if canceled {
		return nil, true, nil
	}
	return idxs, false, nil
}
