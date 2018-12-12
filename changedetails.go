package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	slack "github.com/ashwanthkumar/slack-go-webhook"
	resty "gopkg.in/resty.v1"
)

type tcResponseDetails struct {
	ID       int    `json:"id"`
	Version  string `json:"version"`
	Username string `json:"username"`
	Date     string `json:"date"`
	Href     string `json:"href"`
	WebURL   string `json:"webUrl"`
	Comment  string `json:"comment"`
	Files    struct {
		Count int `json:"count"`
		File  []struct {
			BeforeRevision string `json:"before-revision"`
			AfterRevision  string `json:"after-revision"`
			ChangeType     string `json:"changeType"`
			File           string `json:"file"`
			RelativeFile   string `json:"relative-file"`
		} `json:"file"`
	} `json:"files"`
	VcsRootInstance struct {
		ID        string `json:"id"`
		VcsRootID string `json:"vcs-root-id"`
		Name      string `json:"name"`
		Href      string `json:"href"`
	} `json:"vcsRootInstance"`
}

// Details contain user, comment and version of a change
type Details struct {
	user    string
	comment string
	version string
}

type tcResponse struct {
	Href   string `json:"href"`
	Change []struct {
		ID       int    `json:"id"`
		Version  string `json:"version"`
		Username string `json:"username"`
		Date     string `json:"date"`
		Href     string `json:"href"`
		WebURL   string `json:"webUrl"`
	} `json:"change"`
	Count int `json:"count"`
}

// ChangeDetails gathers details about the given changes from the TeamCity REST api
func ChangeDetails(tcServer, changeurl string) (change Details) {

	tcChangeurl := fmt.Sprint(tcServer, changeurl)

	resp, err := resty.R().
		SetResult(tcResponseDetails{}).
		SetHeader("Accept", "application/json").
		Get(tcChangeurl)

	if err != nil {
		return
	}

	respBody := resp.Result().(*tcResponseDetails)
	change = Details{respBody.Username, respBody.Comment, respBody.Version}

	return change

}

// postChanges posts the collected changes to a slack channel
func postChanges(buildid int) (status string, err error) {

	tcServer := fmt.Sprintf("http://%v:%v@%v", os.Getenv("TCUSER"), os.Getenv("TCPASS", os.Getenv("TCHOST")))
	tcPath := fmt.Sprintf("/app/rest/changes?locator=build:(id:%d)", buildid)
	tcChangeurl := fmt.Sprint(tcServer, tcPath)
	resp, err := resty.R().
		SetResult(tcResponse{}).
		SetHeader("Accept", "application/json").
		Get(tcChangeurl)

	if err != nil {
		return
	}
	respBody := resp.Result().(*tcResponse)

	changeMessage := make(map[int]Details)
	for i := range respBody.Change {
		_, changeMessage[i] = changeMessage, ChangeDetails(tcServer, respBody.Change[i].Href)
	}

	var b strings.Builder
	for i := range changeMessage {
		fmt.Fprintf(&b, "---\nUser: %v\nChange: %v\nVersion: %v\n\n", changeMessage[i].user, changeMessage[i].comment, changeMessage[i].version)
	}
	textMessage := fmt.Sprint(b.String())

	if err != nil {
		return
	}

	webhookURL := os.Getenv("SLACKWEBHOOKURL")
	payload := slack.Payload{
		Text:     textMessage,
		Username: "tcChanges",
	}

	slack.Send(webhookURL, "", payload)

	return resp.Status(), err

}

func main() {
	buildid, _ := strconv.Atoi(os.Getenv("TCBUILDID"))
	_, err := postChanges(buildid)
	if err != nil {
		fmt.Println("Something went wrong")
	}
	fmt.Printf("Ran with BuildID: %v, SlackURL: %v, User: %v, Pass: %v\n", buildid, os.Getenv("SLACKWEBHOOKURL"), os.Getenv("TCUSER"), os.Getenv("TCPASS"))

}
