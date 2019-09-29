package cli

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

// courtesy of : https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
func keysFromMap(mymap map[string]string) []string {

	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

func mergeJiraIssues(src []jira.Issue, dst []jira.Issue) []jira.Issue {
	ticketIds := make(map[string]bool)
	for _, v := range dst {
		ticketIds[v.ID] = true
	}
	for _, v := range src {
		if _, prs := ticketIds[v.ID]; !prs {
			dst = append(dst, v)
		}
	}

	return dst
}

func canonicalAction(action IssueActionBase) string {
	return action.Key() + " " + strings.Join(action.ToParams(), " ")
}

func CancelError() error {
	return errors.New("Cancelled")
}

func IsCancelError(err error) bool {
	return strings.Index(err.Error(), "Cancelled") >= 0
}

func NewRateLimiter(rate time.Duration, burstLimit int) chan time.Time {
	tick := time.NewTicker(rate)
	throttle := make(chan time.Time, burstLimit)
	go func() {
		for t := range tick.C {
			select {
			case throttle <- t:
			default:
			}
		} // does not exit after tick.Stop()
	}()
	return throttle
}
