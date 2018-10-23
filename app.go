package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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

func GithubEventHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	// PullRequestManual(ctx)

	webhook := Webhook{}
	webhook.EventType = r.Header.Get("X-GitHub-Event")
	payload, err := github.ValidatePayload(r, []byte(secretGithub))

	if err != nil {
		panic("Invalid signature")
	}

	webhook.Payload = payload

	GetGithubComment(&webhook, ctx)

	// opt := *github.RepositoryListOptions{}
	// opt := github2.RepositoriesOption

	// RepositoryListOptions{Sort: "created"}

	// opt :=
	// .RepositoryListOptions{
	// 	ListOptions: github.ListOptions{PerPage: 100},
	// }

	// repos, _, err := github2.Repositories.List(ctx, *user.Login, github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 100}})

	// if err == nil {
	// 	return
	// }

	// for _, repo := range repos {
	// 	log.Infof(ctx, "info: repo name: %v\n", repo.Name)
	// }

	conf, err := ParseFile("./github-config.json")
	if err != nil {
		panic("Invalid Config")
	}

	log.Infof(ctx, "StateInfo: %s", webhook.State)
	log.Infof(ctx, "StateInfo: %s", webhook.OriginComment)

	comment := ReplaceComment(webhook.OriginComment, conf)

	log.Infof(ctx, "StateInfo: %s", comment)

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

// DocBaseEvent Hook function
func DocBaseEventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	reportDate := time.Now().Add(time.Hour * 24)
	log.Infof(ctx, "Info: %s", reportDate)

	if r.Header.Get("User-Agent") != "DocBase Webhook/1.0" {
		log.Errorf(ctx, "Invalid User-Agent.", http.StatusInternalServerError)
		return
	}

	var d Docbase // or var d interface{}
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		log.Errorf(ctx, "Err: %s", err.Error())
	}
	log.Infof(ctx, "Info: receive data %s", d)
	log.Infof(ctx, "Info: receive start")
	log.Infof(ctx, "Info: receive data %s", d.Comment.User.Name)
	log.Infof(ctx, "Info: receive data %s", d.Comment.Body)
	conf, err := ParseFile("./docbase-config.json")
	if err != nil {
		log.Errorf(ctx, "Err: %s", err.Error())
	}

	text := ""
	switch d.Action {
	// please add case if you add more action
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
			// w.WriteJson(fmt.Sprintf(`{"res": "%v", "error": "%v"}`, res, err))
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
