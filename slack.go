package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"google.golang.org/appengine/urlfetch"
)

var slackWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")

func sendToSlack(ctx context.Context, text string) (string, error) {
	data := url.Values{}
	data.Set("payload", "{\"text\": \""+text+"\", \"link_names\": 1}")

	client := urlfetch.Client(ctx)
	req, _ := http.NewRequest("POST", slackWebhookURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)

	fmt.Print(res.Body)

	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	if err != nil {
		return string(b), err
	}
	return string(b), nil
}
