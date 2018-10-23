package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Config is ...
type Config struct {
	Accounts map[string]string `json:"accounts"`
}

func init() {
	http.HandleFunc("/github/events", GithubEventHandler)
	http.HandleFunc("/docbase/events", DocBaseEventHandler)
}

// GithubEventHandler is ...
func GithubEventHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)
	webhook := Webhook{}
	webhook.EventType = r.Header.Get("X-GitHub-Event")
	payload, err := github.ValidatePayload(r, []byte(secretGithub))

	if err != nil {
		panic("Invalid signature")
	}

	webhook.Payload = payload

	GetGithubComment(&webhook, ctx)

	conf, err := ParseFile("./github-config.json")
	if err != nil {
		panic("Invalid Config")
	}

	comment := ReplaceComment(webhook.OriginComment, conf)

	if comment == webhook.OriginComment {
		return
	}

	var text string
	text = fmt.Sprintf("%v *【%v】%v* \n", text, webhook.Repository, webhook.Title)
	text = fmt.Sprintf("%v%v\n", text, webhook.HTMLURL)
	text = fmt.Sprintf("%v>Comment created by: %v\n", text, webhook.User)
	text = fmt.Sprintf("%v\n%v\n", text, comment)

	sendToSlack(ctx, text)
}

// DocBaseEventHandler ...
func DocBaseEventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Header.Get("User-Agent") != "DocBase Webhook/1.0" {
		log.Errorf(ctx, "Invalid User-Agent.", http.StatusInternalServerError)
		return
	}

	var d Docbase
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		log.Errorf(ctx, "Err: %s", err.Error())
	}
	conf, err := ParseFile("./docbase-config.json")
	if err != nil {
		log.Errorf(ctx, "Err: %s", err.Error())
	}

	text := ""
	switch d.Action {

	case "comment_create":
		replacedText, isReplaced := ReplaceText(d.Comment.Body, conf)
		if !isReplaced {
			text = ""
		} else {
			text = fmt.Sprintf("%v *【%v】* \n", text, d.Post.Title)
			text = fmt.Sprintf("%v%v\n", text, d.Post.URL)
			text = fmt.Sprintf("%v>Comment created by: %v\n", text, d.Comment.User.Name)
			text = fmt.Sprintf("%v\n%v\n", text, replacedText)
		}
	default:
	}

	if text != "" {
		res, err := sendToSlack(ctx, text)

		if err != nil {
			log.Infof(ctx, "Info: %s", res)
		}
	}
}

// ParseFile is ...
func ParseFile(filename string) (*Config, error) {
	c := Config{}

	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonString, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func initGithubClient(ctx context.Context, secretGithub string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: secretGithub,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client
}
