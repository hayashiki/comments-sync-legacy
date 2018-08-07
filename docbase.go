package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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

var r = regexp.MustCompile(`@[a-zA-Z0-9_\-]+`)

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
