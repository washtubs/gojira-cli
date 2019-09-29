package cli

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/andygrunwald/go-jira"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

type JiraClientFactory struct {
	config Config
}

func (j *JiraClientFactory) GetClient() (*jira.Client, error) {
	config := j.config.Client
	crt, err := tls.LoadX509KeyPair(config.Certfile, config.Keyfile)
	if err != nil {
		log.Println("Error loading X509 pair for cert " + config.Certfile +
			" and key file " + config.Keyfile)
		return nil, err
	}

	passfile, err := os.Open(config.Passfile)
	if err != nil {
		log.Println("Failed to open " + config.Passfile)
		return nil, err
	}
	defer passfile.Close()

	bs, err := ioutil.ReadAll(passfile)
	if err != nil {
		return nil, err
	}
	password := string(bs)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{crt},
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	cl := &http.Client{
		Transport: tr,
	}
	jcli, err := jira.NewClient(cl, config.Url)
	jcli.Authentication.SetBasicAuth(config.Username, password)
	return jcli, err
}

func NewJiraClientFactory() *JiraClientFactory {
	return &JiraClientFactory{}
}

type IssueLinkTypeMenu struct {
	jiraClientFactory *JiraClientFactory
	issueLinkTypes    []jira.IssueLinkType
	cursor            int
}

func (m *IssueLinkTypeMenu) IssueLinkType() (linkType jira.IssueLinkType, subjectIsInward bool) {
	linkType = m.issueLinkTypes[m.cursor/2]
	subjectIsInward = m.cursor%2 == 0
	return
}

func (m *IssueLinkTypeMenu) Select(subjectIssueId string) error {
	if m.issueLinkTypes == nil {
		client, err := m.jiraClientFactory.GetClient()
		if err != nil {
			return err
		}

		apiEndpoint := fmt.Sprintf("rest/api/2/issueLinkType")
		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return err
		}

		var issueLinkTypesWrap map[string][]jira.IssueLinkType
		resp, err := client.Do(req, issueLinkTypesWrap)
		LogHttpResponse(resp)
		if err != nil {
			return err
		}
		var prs bool
		m.issueLinkTypes, prs = issueLinkTypesWrap["issueLinkTypes"]
		if !prs {
			return errors.New("Unexpected format, no issueLinkTypes key")
		}
	}

	labels := make([]string, len(m.issueLinkTypes)*2)
	for i, link := range m.issueLinkTypes {
		labels[i*2] = subjectIssueId + " - " + link.Inward + "..."
		labels[i*2+1] = subjectIssueId + " - " + link.Outward + "..."
	}

	p := promptui.Select{
		Label: "Choose a relation",
		Items: labels,
		Size:  10,
	}
	cursor, _, err := p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: ", err)
		return err
	}

	m.cursor = cursor
	return nil
}
