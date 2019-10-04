package cli

import (
	"log"

	"github.com/andygrunwald/go-jira"
)

func LogHttpResponse(resp *jira.Response) {
	if resp == nil {
		return
	}
	if resp.StatusCode >= 400 {
		log.Printf("HTTP error code=[%d | %s] %s", resp.StatusCode, resp.Status, resp.Response.Request.URL.String())
	}
}
