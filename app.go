package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// Config is ...
type Config struct {
	Accounts map[string]string `json:"accounts"`
}

func init() {
	http.HandleFunc("/docbase/events", DocBaseEventHandler)
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

	conf, err := ParseFile("./config.json")
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
			text = fmt.Sprintf("%v>Comment created by: %v\n", text, d.Post.User.Name)
			text = fmt.Sprintf("%v\n%v\n", text, replacedText)
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
