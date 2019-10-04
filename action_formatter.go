package cli

import (
	"bytes"
	"log"
	"strings"
	"text/template"
)

// Formats an action potentially in conjunction with an issue
type IssueActionFormatter struct {
	IssueAction
}

func (f IssueActionFormatter) Format() string {
	issueRef := "an issue"
	if f.issue.Key != "" {
		issueRef = f.issue.Key
	}
	t := strings.ReplaceAll(f.action.Template(), "_ISSUE", issueRef)

	tpl, err := template.New("").Parse(t)
	if err != nil {
		log.Printf("Error parsing template: %s", err.Error())
		return t
	}

	buf := bytes.NewBufferString("")
	err = tpl.Execute(buf, f.IssueAction.action)
	if err != nil {
		log.Printf("Error executing template: %s", err.Error())
		return t
	}
	return buf.String()
}

type ActionTypeFormatter struct {
	IssueActionBase
}

func (f ActionTypeFormatter) Format() string { return f.Description() }

func WrapFormatter(action IssueActionBase) Formatter {
	if !action.IsBuilt() {
		return ActionTypeFormatter{action}
	}

	return IssueActionFormatter{IssueAction{action: action}}

}
