package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/appengine/urlfetch"
)

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
