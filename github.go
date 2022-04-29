package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"google.golang.org/appengine/log"
)

type Webhook struct {
	EventType       string
	Payload         []byte
	Repository      string
	Title           string
	OriginComment   string
	ReplacedComment string
	User            string
	HTMLURL         string
	State           string
}

// Repository is
type Repository struct {
	Owner string
	Name  string
}

// File is
type File struct {
	RawReviewers []interface{} `json:"reviewers"`
}

var secretGithub = os.Getenv("GITHUB_SECRET")

// GetGithubComment is ...
func GetGithubComment(webhook *Webhook, ctx context.Context) {
	switch webhook.EventType {
	case "issue_comment":
		IssueComment(webhook, ctx)
	case "pull_request_review_comment":
		PullRequestComment(webhook, ctx)
	case "pull_request_review":
		PullRequestReview(webhook, ctx)
	case "push":
		Push(webhook, ctx)
	default:
		log.Infof(ctx, "info: Event doesn't exist")
	}
}

// Push is ...
func Push(webhook *Webhook, ctx context.Context) {
	var pushEventGithub github.PushEvent
	err := json.Unmarshal(webhook.Payload, &pushEventGithub)
	if err != nil {
		panic(err)
	}
}

// PullRequestReview is ...
func PullRequestReview(webhook *Webhook, ctx context.Context) {
	log.Infof(ctx, "debug: run PullRequestReview")

	var pullrequestReviewGithub github.PullRequestReviewEvent
	err := json.Unmarshal(webhook.Payload, &pullrequestReviewGithub)
	if err != nil {
		panic(err)
	}

	if *pullrequestReviewGithub.Review.State == "approved" {

		var secretGithub = os.Getenv("GITHUB_SECRET_TOKEN")
		client := initGithubClient(ctx, secretGithub)

		repoOwner := *pullrequestReviewGithub.Repo.Owner.Login
		repo := *pullrequestReviewGithub.Repo.Name
		reviewers, _, _ := client.PullRequests.ListReviewers(ctx, repoOwner, repo, *pullrequestReviewGithub.PullRequest.Number, &github.ListOptions{PerPage: 100})
		reviews, _, _ := client.PullRequests.ListReviews(ctx, repoOwner, repo, *pullrequestReviewGithub.PullRequest.Number, &github.ListOptions{PerPage: 100})
		reviewsName := []string{}
		for _, review := range reviews {
			if *review.State == "APPROVED" {
				reviewsName = append(reviewsName, *review.User.Login)
			}
		}

		reviewersName := []string{}
		for _, reviewer := range reviewers.Users {
			reviewersName = append(reviewersName, *reviewer.Login)
		}

		webhook.Repository = *pullrequestReviewGithub.Repo.Name
		webhook.User = *pullrequestReviewGithub.Sender.Login
		webhook.Title = *pullrequestReviewGithub.PullRequest.Title
		webhook.OriginComment = fmt.Sprintf("%v%v%v%v%v%v", ":white_check_mark: @", *pullrequestReviewGithub.PullRequest.User.Login, "approved", reviewsName, "not approved", reviewersName)
		webhook.HTMLURL = *pullrequestReviewGithub.PullRequest.HTMLURL
		webhook.State = *pullrequestReviewGithub.Review.State
	}
}

func IssueComment(webhook *Webhook, ctx context.Context) {
	var issueGithub github.IssueCommentEvent
	err := json.Unmarshal(webhook.Payload, &issueGithub)
	if err != nil {
		panic(err)
	}

	webhook.Repository = *issueGithub.Repo.Name
	webhook.Title = *issueGithub.Issue.Title
	webhook.User = *issueGithub.Comment.User.Login
	webhook.OriginComment = *issueGithub.Comment.Body
	webhook.HTMLURL = *issueGithub.Comment.HTMLURL

	var secretGithub = os.Getenv("GITHUB_SECRET_TOKEN")
	client := initGithubClient(ctx, secretGithub)

	repoOwner := *issueGithub.Repo.Owner.Login
	log.Infof(ctx, "debug: repository owner is %v\n", repoOwner)
	repo := *issueGithub.Repo.Name
	log.Infof(ctx, "debug: repository name is %v\n", repo)

	issueSvc := client.Issues

	// issue := *issueGithub.Issue
	issueNum := *issueGithub.Issue.Number

	// コメントにr？ が含まれる場合
	if string([]rune(webhook.OriginComment)[:2]) == "r?" {
		ReviewerList := getReviewersFromConf(ctx, *issueGithub.Issue.User.Login)
		ReviewerListWithAt := getReviewersFromConfWithAt(ctx, *issueGithub.Issue.User.Login)
		Reviewers := github.ReviewersRequest{Reviewers: ReviewerList}

		comment := new(github.IssueComment)
		comment.Body = github.String(strings.Join(ReviewerListWithAt, ",") + "レビューお願いします")
		issueSvc.EditComment(ctx, repoOwner, repo, *issueGithub.Comment.ID, comment)
		client.PullRequests.RequestReviewers(ctx, repoOwner, repo, issueNum, Reviewers)
	}
}

func changeStatusLabel(list []*github.Label, new string) []string {
	result := make([]string, 0, 0)
	for _, item := range list {
		label := *item.Name
		result = append(result, label)
	}
	result = append(result, new)
	return result
}

// PullRequestComment is ...
func PullRequestComment(webhook *Webhook, ctx context.Context) {
	var pullrequestGithub github.PullRequestReviewCommentEvent
	err := json.Unmarshal(webhook.Payload, &pullrequestGithub)
	if err != nil {
		panic(err)
	}

	webhook.Repository = *pullrequestGithub.Repo.Name
	webhook.User = *pullrequestGithub.Comment.User.Login
	webhook.Title = *pullrequestGithub.PullRequest.Title
	webhook.OriginComment = *pullrequestGithub.Comment.Body
	webhook.HTMLURL = *pullrequestGithub.Comment.HTMLURL

	var secretGithub = os.Getenv("GITHUB_SECRET_TOKEN")
	client := initGithubClient(ctx, secretGithub)
	repoOwner := *pullrequestGithub.Repo.Owner.Login
	log.Infof(ctx, "debug: repository owner is %v\n", repoOwner)
	repo := *pullrequestGithub.Repo.Name
	log.Infof(ctx, "debug: repository name is %v\n", repo)

	opt := &github.ListOptions{}
	reviews, resp, err := client.PullRequests.ListReviewers(ctx, repoOwner, repo, *pullrequestGithub.PullRequest.Number, opt)

	log.Infof(ctx, "debug: reviews %v\n", reviews)
	log.Infof(ctx, "debug: resp %v\n", resp)

}

// ReplaceComment replace github account to slack
func ReplaceComment(comment string, conf *Config) string {

	matches := r.FindAllStringSubmatch(comment, -1)
	for _, val := range matches {
		slackName, _ := conf.Accounts[val[0]]
		comment = strings.Replace(comment, val[0], slackName, -1)
	}
	return comment
}

func getReviewersFromConf(ctx context.Context, me string) []string {
	bytes, err := ioutil.ReadFile("github-config.json")

	if err != nil {
		log.Errorf(ctx, "error: resp %v\n", err)
	}

	var file File

	if err := json.Unmarshal(bytes, &file); err != nil {
		log.Errorf(ctx, "error: resp %v\n", err)
	}

	var list []string
	for _, r := range file.RawReviewers {

		if me == r.(string) {
			log.Errorf(ctx, "info: me %v\n", me)
		} else {
			list = append(list, r.(string))
		}
	}
	return list
}

func getReviewersFromConfWithAt(ctx context.Context, me string) []string {
	bytes, err := ioutil.ReadFile("test.json")

	if err != nil {
		log.Errorf(ctx, "error: resp %v\n", err)
	}

	var file File

	if err := json.Unmarshal(bytes, &file); err != nil {
		log.Errorf(ctx, "error: resp %v\n", err)
	}

	var list []string
	for _, r := range file.RawReviewers {

		if me == r.(string) {
			log.Errorf(ctx, "info: me %v\n", me)
		} else {
			list = append(list, "@"+r.(string))
		}
	}
	return list
}
