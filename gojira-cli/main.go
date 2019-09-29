package main

import (
	"flag"
	"log"

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
		case "print":
			issueId := flag.Arg(2)
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
