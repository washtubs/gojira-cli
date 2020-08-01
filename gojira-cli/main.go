package main

import (
	"flag"
	"log"
	"strings"

	cli "github.com/washtubs/gojira-cli"
)

func main() {
	flag.Parse()
	if flag.Arg(0) == "_rpc" {
		c := cli.NewRpcClient()
		action := flag.Arg(1)
		switch action {
		case "load":
			c.LoadResults()
		case "query":
			fifo := flag.Arg(2)
			if fifo == "" {
				log.Fatal("Expected a fifo that the rpc server can access as the first argument")
			}
			c.Query(fifo)
		case "print":
			fzfRecord := flag.Arg(2)
			fields := strings.Fields(fzfRecord)
			if len(fields) < 2 {
				log.Fatal("Expecting issueId as the second field. Not enough fields")
			}
			issueId := fields[1]
			if issueId == "" {
				log.Fatal("Must provide an issueId")
			}
			c.PrintIssue(issueId)
		default:
			log.Fatal("Unknown action " + action)
		}
		return
	}

	cli.RunWorkbench()
}
