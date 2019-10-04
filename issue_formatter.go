package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/manifoldco/promptui"
)

type Formatter interface {
	Format() string
}

type IssueFormattable struct {
	issue     jira.Issue
	formatter IssueFormatter
}

func (f IssueFormattable) Format() string {
	return f.formatter.Format(f.issue)
}

type IssueFormatter interface {
	ExtractTicketId(formatted string) string
	Format(issue jira.Issue) string
}

type FormatterConfig struct {
	excludeSummary  bool
	includeLabels   bool
	includeReporter bool
}

var formatterFlags = []string{
	"summary",
	"labels",
	"reporter",
}

type FormatterMenu struct {
	*FormatterConfig
	cursor int
}

func (m *FormatterMenu) Get(flag string) bool {
	switch flag {
	case "summary":
		return !m.excludeSummary
	case "labels":
		return m.includeLabels
	case "reporter":
		return m.includeReporter
	default:
		panic("unhandled flag " + flag)
	}
}

func (m *FormatterMenu) Toggle(flag string) {
	switch flag {
	case "summary":
		m.excludeSummary = !m.excludeSummary
	case "labels":
		m.includeLabels = !m.includeLabels
	case "reporter":
		m.includeReporter = !m.includeReporter
	default:
		panic("unhandled flag " + flag)
	}
}

func (m *FormatterMenu) Select() error {
	options := make([]string, 0)
	for _, flag := range formatterFlags {
		label := "Include "
		if !m.Get(flag) {
			label = "Exclude "
		}
		options = append(options, label+flag)
	}

	p := promptui.Select{
		Label: "Toggle issue format options",
		Items: options,
		Size:  10,
	}

	cursor, _, err := p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: ", err)
		return err
	}
	m.cursor = cursor

	toggleFlag := formatterFlags[cursor]
	m.Toggle(toggleFlag)
	return nil
}

type defaultIssueFormatter struct {
	*FormatterConfig
}

func (f *defaultIssueFormatter) ExtractTicketId(formatted string) string {
	split := strings.Split(formatted, "-")
	return strings.TrimSpace(split[0])
}

func (f *defaultIssueFormatter) Format(issue jira.Issue) string {

	out := issue.ID

	out = out + " " + issue.Key + " -"

	if !f.excludeSummary {
		out = out + " " + issue.Fields.Summary
	}

	if f.includeLabels {
		for _, label := range issue.Fields.Labels {
			out = out + " " + "[" + label + "]"
		}
	}

	if f.includeReporter && issue.Fields.Reporter != nil {
		out = out + issue.Fields.Reporter.Name
	}

	return out
}

func PrintIssue(issue jira.Issue) string {
	status := ""
	if issue.Fields.Status != nil {
		status = issue.Fields.Status.Name
	}
	return fmt.Sprintf(`%s: %s
Description: %s
Status: %s
`,
		issue.Key,
		issue.Fields.Summary,
		issue.Fields.Description,
		status,
	)
}

func NewIssueFormatter(config *FormatterConfig) IssueFormatter {
	return &defaultIssueFormatter{config}
}
