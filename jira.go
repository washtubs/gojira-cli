package cli

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

type JiraClientFactory struct {
	config *Config
	client *jira.Client
}

func (j *JiraClientFactory) GetClient() (*jira.Client, error) {
	if j.client != nil {
		return j.client, nil
	}
	config := j.config.Client
	crt, err := tls.LoadX509KeyPair(config.Certfile, config.Keyfile)
	if err != nil {
		log.Println("Error loading X509 pair for cert " + config.Certfile +
			" and key file " + config.Keyfile + " : " + err.Error())
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
	password := strings.TrimSpace(string(bs))

	tlsConfig := &tls.Config{
		Certificates:  []tls.Certificate{crt},
		Renegotiation: tls.RenegotiateOnceAsClient,
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	cl := &http.Client{
		Transport: tr,
	}
	j.client, err = jira.NewClient(cl, config.Url)
	j.client.Authentication.SetBasicAuth(config.Username, password)

	return j.client, err
}

func NewJiraClientFactory(config *Config) *JiraClientFactory {
	return &JiraClientFactory{config: config}
}

type IssueLinkTypeMenu struct {
	jiraClientFactory *JiraClientFactory
	issueLinkTypes    []jira.IssueLinkType
	cursor            int
}

func (m *IssueLinkTypeMenu) IssueLinkType() (linkType jira.IssueLinkType, subjectIsInward bool) {
	linkType = m.issueLinkTypes[m.cursor/2]
	subjectIsInward = m.cursor%2 == 1
	return
}

func (m *IssueLinkTypeMenu) Select(subjectIssue jira.Issue) error {
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
		resp, err := client.Do(req, &issueLinkTypesWrap)
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
		labels[i*2] = subjectIssue.Key + " - " + link.Inward + "..."
		labels[i*2+1] = subjectIssue.Key + " - " + link.Outward + "..."
	}

	p := promptui.Select{
		Label: "Choose a relation",
		Items: labels,
		Size:  10,
	}
	cursor, _, err := p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: %s", err)
		return err
	}

	m.cursor = cursor
	return nil
}
