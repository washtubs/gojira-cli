package cli

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

type SearchInteractor interface {
	Loaded() []jira.Issue
	Append(issue jira.Issue)
	LoadResults()
	CloseSearch()
}

type defaultSearchInteractor struct {
	loadChannel   chan bool
	cancelChannel chan bool
	pageSize      int
	closed        bool

	mutex  sync.Mutex
	loaded []jira.Issue
}

func (is *defaultSearchInteractor) Loaded() []jira.Issue {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	return is.loaded
}

func (is *defaultSearchInteractor) Append(issue jira.Issue) {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	is.loaded = append(is.loaded, issue)
}

func (is *defaultSearchInteractor) LoadResults() {
	if is.closed {
		return
	}
	for i := 0; i < is.pageSize; i++ {
		select {
		case is.loadChannel <- true:
		default:
			break
		}
	}
}

func (is *defaultSearchInteractor) CloseSearch() {
	if is.closed {
		return
	}
	select {
	case is.cancelChannel <- true:
	default:
	}
	is.closed = true
}

type IssueEnumerator interface {
	ForEachIssue(jql string, opts *jira.SearchOptions, each func(jira.Issue) error) error
}

type jiraIssueEnum struct {
	jiraClientFactory *JiraClientFactory
}

func (j *jiraIssueEnum) ForEachIssue(jql string, opts *jira.SearchOptions, each func(jira.Issue) error) error {
	client, err := j.jiraClientFactory.GetClient()
	if err != nil {
		return err
	}

	return client.Issue.SearchPages(jql, opts, each)
}

type IssueSearcher interface {
	SetSearchQuery(jql string)
	SearchAsync() (<-chan jira.Issue, SearchInteractor)
}

// Defined in jira.SearchOptions
const defaultStubPageSize = 10

func getMaxResults(opts *jira.SearchOptions) int {
	if opts.MaxResults == 0 {
		return 50
	}
	return opts.MaxResults
}

type defaultIssueSearcher struct {
	enumerator IssueEnumerator

	// mutex guards access to all below fields
	mutex sync.Mutex
	opts  *jira.SearchOptions
	jql   string
}

func (is *defaultIssueSearcher) SetSearchQuery(jql string) {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	// Reset everything
	is.opts.StartAt = 0
	is.jql = jql
}

func (is *defaultIssueSearcher) SearchAsync() (<-chan jira.Issue, SearchInteractor) {

	is.mutex.Lock()
	defer is.mutex.Unlock()

	loadChannel := make(chan bool, getMaxResults(is.opts))
	cancelChannel := make(chan bool, 1)
	issueChan := make(chan jira.Issue, getMaxResults(is.opts))

	interactor := &defaultSearchInteractor{
		loadChannel:   loadChannel,
		cancelChannel: cancelChannel,
		pageSize:      getMaxResults(is.opts),
	}

	interactor.loaded = make([]jira.Issue, 0, getMaxResults(is.opts)*2) // no alloc until after 2 pages of results
	go func() {
		defer func() {
			interactor.closed = true
			close(loadChannel)
			close(cancelChannel)
			close(issueChan)
		}()
		done := false
		err := is.enumerator.ForEachIssue(is.jql, is.opts, func(issue jira.Issue) error {
			select {
			case <-cancelChannel:
				return errors.New("Done")
			case <-loadChannel:
			}

			issueChan <- issue
			interactor.Append(issue)
			return nil
		})
		if !done && err != nil {
			log.Println("Error while searching: " + err.Error())
		}
	}()

	interactor.LoadResults()

	return issueChan, interactor
}

func makeFakeIssue(project string, id int, summary string) jira.Issue {
	issue := jira.Issue{}
	issue.Fields = &jira.IssueFields{}
	issue.ID = project + "-" + strconv.Itoa(id)
	issue.Fields.Summary = summary
	issue.Fields.Status = &jira.Status{
		Name: "In Progress",
	}
	return issue
}

type mockIssueEnum struct {
	issues []jira.Issue
}

func (is *mockIssueEnum) ForEachIssue(jql string, opts *jira.SearchOptions, each func(jira.Issue) error) error {
	issues := make([]jira.Issue, 0)
	for _, v := range is.issues {
		if strings.Index(jql, "FOO") >= 0 {
			if strings.Index(v.ID, "FOO") == 0 {
				issues = append(issues, v)
			}
		} else if strings.Index(jql, "BAR") >= 0 {
			if strings.Index(v.ID, "BAR") == 0 {
				issues = append(issues, v)
			}
		} else {
			issues = append(issues, v)
		}
	}
	for _, issue := range issues {
		err := each(issue)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: convert to real jira searcher
func InitIssueSearcher(jiraClientFactory *JiraClientFactory) IssueSearcher {
	return &defaultIssueSearcher{
		opts:       &jira.SearchOptions{},
		jql:        "",
		mutex:      sync.Mutex{},
		enumerator: &jiraIssueEnum{jiraClientFactory},
	}
}
