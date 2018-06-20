package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Docbase struct {
	Trigger struct {
		Name  string `json:"name"`
		Label string `json:"label"`
	} `json:"trigger"`
	Message string `json:"message"`
	Action  string `json:"action"`
	Sender  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"sender"`
	Team struct {
		Domain string `json:"domain"`
		Name   string `json:"name"`
	} `json:"team"`
	Post struct {
		ID        int       `json:"id"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		Draft     bool      `json:"draft"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"created_at"`
		Tags      []struct {
			Name string `json:"name"`
		} `json:"tags"`
		Scope  string `json:"scope"`
		Groups []struct {
			Name string `json:"name"`
		} `json:"groups"`
		User struct {
			ID              int    `json:"id"`
			Name            string `json:"name"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"user"`
	} `json:"post"`
	Comment struct {
		ID        int       `json:"id"`
		Body      string    `json:"body"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"comment"`
	Users []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		ProfileImageURL string `json:"profile_image_url"`
	} `json:"users"`
}

// Config is ...
type Config struct {
	Accounts map[string]string `json:"accounts"`
}

var r = regexp.MustCompile(`@[a-zA-Z0-9_\-]+`)

func init() {
	http.HandleFunc("/docbase/events", PostDocBaseEvents)
}

// DocBaseEvent Hook function
func PostDocBaseEvents(w http.ResponseWriter, r *http.Request) {
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
	log.Infof(ctx, "Info: %s", d)

	conf, err := ParseFile("./config.json")
	if err != nil {
		log.Errorf(ctx, "Err: %s", err.Error())
	}

	text := ""
	switch d.Action {
	// please add case if you add more action
	case "comment_create":
		replacedText, isReplaced := ReplaceText(d.Message, conf)
		if !isReplaced {
			text = ""
		} else {
			text = replacedText
		}
	default:
	}

	if text != "" {
		endPoint := os.Getenv("SLACK_INCOMING_WEBHOOK")
		res, err := SendToSlack(ctx, fmt.Sprintf("/services/%v", endPoint), text)

		if err != nil {
			log.Infof(ctx, "Info: %s", res)
			// w.WriteJson(fmt.Sprintf(`{"res": "%v", "error": "%v"}`, res, err))
		}
	}
}

func SendToSlack(ctx context.Context, path string, text string) (string, error) {

	slackURL := "https://hooks.slack.com"
	slackPath := path
	u, _ := url.ParseRequestURI(slackURL)
	u.Path = slackPath

	urlStr := fmt.Sprintf("%v", u)

	data := url.Values{}
	data.Set("payload", "{\"text\": \""+text+"\", \"link_names\": 1}")

	client := urlfetch.Client(ctx)
	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)

	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	if err != nil {
		return string(b), err
	}
	return string(b), nil
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

func ReplaceText(text string, conf *Config) (string, bool) {
	matches := r.FindAllStringSubmatch(text, -1)
	replaceFlg := false
	fmt.Println(matches)
	for _, val := range matches {
		slackName, _ := conf.Accounts[val[0]]
		fmt.Println(slackName)
		text = strings.Replace(text, val[0], slackName, -1)
		replaceFlg = true
	}
	return text, replaceFlg
}
