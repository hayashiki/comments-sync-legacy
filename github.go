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

// type PullRequestReview struct {
// 	ID             *int       `json:"id,omitempty"`
// 	User           *User      `json:"user,omitempty"`
// 	Body           *string    `json:"body,omitempty"`
// 	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
// 	CommitID       *string    `json:"commit_id,omitempty"`
// 	HTMLURL        *string    `json:"html_url,omitempty"`
// 	PullRequestURL *string    `json:"pull_request_url,omitempty"`
// 	State          *string    `json:"state,omitempty"`
// }

var secretGithub = os.Getenv("GITHUB_SECRET")

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

func Push(webhook *Webhook, ctx context.Context) {
	log.Infof(ctx, "debug: run Push")
	var pushEventGithub github.PushEvent
	err := json.Unmarshal(webhook.Payload, &pushEventGithub)
	if err != nil {
		panic(err)
	}
	log.Infof(ctx, "debug: run Push %v", *pushEventGithub.Repo.Name)
	log.Infof(ctx, "debug: run Push %v", *pushEventGithub.After)
	log.Infof(ctx, "debug: run Push %v", *pushEventGithub.Ref)
	log.Infof(ctx, "debug: run Push %v", *pushEventGithub.Sender.Name)

	// https://github.com/framgia/ARRANGE_Server/commit/c59da5535df206a218432da3cacf4da08448c15f

	// webhook.Repository = *pushEventGithub.Repo.Name
	// webhook.User = *pushEventGithub.Sender.Name
	// webhook.Title = *pullrequestReviewGithub.PullRequest.Title
	// webhook.OriginComment = fmt.Sprintf("%v%v%v%v%v%v", ":white_check_mark: @", *pullrequestReviewGithub.PullRequest.User.Login, "approved", reviewsName, "not approved", reviewersName)
	// webhook.HTMLURL = *pullrequestReviewGithub.PullRequest.HTMLURL
}

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
		log.Infof(ctx, "debug: repository owner is %v\n", repoOwner)
		repo := *pullrequestReviewGithub.Repo.Name
		log.Infof(ctx, "debug: repository name is %v\n", repo)

		reviewers, _, _ := client.PullRequests.ListReviewers(ctx, repoOwner, repo, *pullrequestReviewGithub.PullRequest.Number, &github.ListOptions{PerPage: 100})
		reviews, _, _ := client.PullRequests.ListReviews(ctx, repoOwner, repo, *pullrequestReviewGithub.PullRequest.Number, &github.ListOptions{PerPage: 100})

		log.Infof(ctx, "debug: num %v\n", *pullrequestReviewGithub.PullRequest.Number)

		reviewsName := []string{}
		for _, review := range reviews {
			log.Infof(ctx, "debug: num %v\n", *review.State)
			if *review.State == "APPROVED" {
				reviewsName = append(reviewsName, *review.User.Login)
			}
		}

		reviewersName := []string{}
		for _, reviewer := range reviewers.Users {
			reviewersName = append(reviewersName, *reviewer.Login)
		}

		// log.Infof(ctx, "debug: reviews %v\n", reviews)
		// log.Infof(ctx, "debug: reviewers %v\n", reviewers)
		log.Infof(ctx, "debug: reviewed %v\n", reviewsName)
		log.Infof(ctx, "debug: not review %v\n", reviewersName)

		webhook.Repository = *pullrequestReviewGithub.Repo.Name
		webhook.User = *pullrequestReviewGithub.Sender.Login
		webhook.Title = *pullrequestReviewGithub.PullRequest.Title
		webhook.OriginComment = fmt.Sprintf("%v%v%v%v%v%v", ":white_check_mark: @", *pullrequestReviewGithub.PullRequest.User.Login, "approved", reviewsName, "not approved", reviewersName)
		webhook.HTMLURL = *pullrequestReviewGithub.PullRequest.HTMLURL
		webhook.State = *pullrequestReviewGithub.Review.State
		// 20951590-ab1e-11e8-89e9-82e1c21520b1

	}
}

func PullRequestManual(ctx context.Context) {

	var secretGithub = os.Getenv("GITHUB_SECRET_TOKEN")
	client := initGithubClient(ctx, secretGithub)

	repoOwner := "framgia"

	repo := "ARRANGE_Server"
	issueNum := 5233

	// issueSvc := client.Issues

	// // 本当にリクエストとぶのでコメ
	// IssueRequest := &github.IssueRequest{
	// 	Title: github.String("CompanyのAPIをsuper_adminとadminで分ける"),
	// 	Body:  github.String(""),
	//	// Labels: &[]string{"radar"},
	// }
	// client.Issues.Create(ctx, repoOwner, repo, IssueRequest)

	reviews, _, _ := client.PullRequests.ListReviews(ctx, repoOwner, repo, issueNum, &github.ListOptions{PerPage: 100})

	// log.Infof(ctx, "debug: num %v\n", reviews)

	reviewsName := []string{}
	for _, review := range reviews {
		log.Infof(ctx, "debug: num %v\n", *review.State)
		if *review.State == "APPROVED" {
			reviewsName = append(reviewsName, *review.User.Login)
		}
	}
	log.Infof(ctx, "debug: num %v\n", reviewsName)

	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "open",
	}

	var apiResponse []*github.Issue

	for {
		issues, resp, err := client.Issues.ListByRepo(ctx, repoOwner, repo, opt)
		if err != nil {
			// return nil, err
			return
		}
		apiResponse = append(apiResponse, issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, issue := range apiResponse {
		log.Infof(ctx, "debug: issue.Title %v\n", *issue)
	}

}

func IssueComment(webhook *Webhook, ctx context.Context) {

	log.Infof(ctx, "debug: run IssueComment")
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

	issue := *issueGithub.Issue
	issueNum := *issueGithub.Issue.Number
	log.Infof(ctx, "debug: issue number is %v\n", issueNum)
	log.Infof(ctx, "debug: issue link is %v\n", issue.PullRequestLinks)

	// コメントにr？ が含まれる場合
	if string([]rune(webhook.OriginComment)[:2]) == "r?" {
		temp1ReviewerList := getReviewersFromConf(ctx, *issueGithub.Issue.User.Login)
		temp3ReviewerList := getReviewersFromConfWithAt(ctx, *issueGithub.Issue.User.Login)
		temp2ReviewerList := []string{"naosk8-zenkigen", "yoshidash23"}
		ReviewerList := github.ReviewersRequest{Reviewers: temp1ReviewerList}

		log.Infof(ctx, "debug: temp1 %v\n", temp1ReviewerList)
		log.Infof(ctx, "debug: temp2 %v\n", temp2ReviewerList)
		log.Infof(ctx, "debug: temp3 %v\n", ReviewerList)
		log.Infof(ctx, "debug: temp5 %v\n", temp3ReviewerList)

		log.Infof(ctx, "debug: temp4 %v\n", *issueGithub.Comment.ID)

		n := int(*issueGithub.Comment.ID)
		// *issueGithub.Comment.ID
		// コメントを消す
		// issueSvc.DeleteComment(ctx, repoOwner, repo, n)

		comment := new(github.IssueComment)
		comment.Body = github.String(strings.Join(temp3ReviewerList, ",") + "レビューお願いします")
		issueSvc.EditComment(ctx, repoOwner, repo, n, comment)
		// issueSvc.CreateComment(ctx, repoOwner, repo, issueNum, comment)
		// 自分以外のレビューをつける
		client.PullRequests.RequestReviewers(ctx, repoOwner, repo, issueNum, ReviewerList)
		return

		// 自分以外の相手にメンションする

		// Comment.idで　DeleteCommentしたい。詳細は以下。
		// そのあと、さらに　CreateComment　する

		// func (s *IssuesService) DeleteComment(ctx context.Context, owner string, repo string, commentID int64) (*Response, error) {
		// 	u := fmt.Sprintf("repos/%v/%v/issues/comments/%d", owner, repo, commentID)
		// 	req, err := s.client.NewRequest("DELETE", u, nil)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	return s.client.Do(ctx, req, nil)
		// }

		// func (s *IssuesService) CreateComment(ctx context.Context, owner string, repo string, number int, comment *IssueComment) (*IssueComment, *Response, error) {
		// 	u := fmt.Sprintf("repos/%v/%v/issues/%d/comments", owner, repo, number)
		// 	req, err := s.client.NewRequest("POST", u, comment)
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	c := new(IssueComment)
		// 	resp, err := s.client.Do(ctx, req, c)
		// 	if err != nil {
		// 		return nil, resp, err
		// 	}

		// 	return c, resp, nil
		// }

	}
	// opt := &github.ListOptions{}

	return

	user, _, err := client.Users.Get(ctx, "")

	if err != nil {
		return
	}

	log.Infof(ctx, "info: %v", user)

	// if github2 == nil {
	// 	log.Infof(ctx, "error: cannot create the github clinet")
	// 	return
	// }

	if *issueGithub.Issue.PullRequestLinks.URL == "" {
		log.Infof(ctx, "info: the issue is pull request %v", *issueGithub.Issue.PullRequestLinks)
	}

	currentLabels := GetLabelsByIssue(ctx, issueSvc, repoOwner, repo, issueNum)
	if currentLabels == nil {
		log.Infof(ctx, "nolabel")
	}

	// AddAwaitingReviewLabel(currentLabels)
	// データ更新系のためコメントアウト
	labels := AddAwaitingReviewLabel(currentLabels)
	_, _, _ = issueSvc.ReplaceLabelsForIssue(ctx, repoOwner, repo, issueNum, labels)

	// zenkigen

	GetRepoPullRequestsAndReportErrors(ctx, client, repoOwner, repo)
	// log.Infof(ctx, "debug: assignees is %v\n", assignees)

	// 自動アサイン機能 データ更新系
	// issueSvc.AddAssignees(ctx, repoOwner, repo, issueNum, []string{"hayashiki", "yoshidash23"})

	// if err != nil {
	// 	log.Infof(ctx, "info: could not change assignees.")
	// 	// return false, err
	// }

}

const (
	LABEL_AWAITING_REVIEW string = "maintenance"
	// LABEL_AWAITING_MERGE            string = "S-awaiting-merge"
	// LABEL_NEEDS_REBASE              string = "S-needs-rebase"
	// LABEL_FAILS_TESTS_WITH_UPSTREAM string = "S-fails-tests-with-upstream"
)

// func RequestReviewers(ctx context.Context, event *eventData, reviewers origin.ReviewersRequest) error {
// 	_, _, err := this.origin.PullRequests.RequestReviewers(
// 		ctx, event.owner, event.repository, event.number, reviewers,
// 	)

// 	return err
// }

func AddAwaitingReviewLabel(list []*github.Label) []string {
	return changeStatusLabel(list, LABEL_AWAITING_REVIEW)
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

// func GetRepoIssues(ctx context.Context, client *github.Client, owner, name string) {

// 	client.Issues.List(ctx, owner, name, opt)
// }

func GetRepoPullRequestsAndReportErrors(ctx context.Context, client *github.Client, owner, name string) {
	log.Infof(ctx, "debug: getRepoPullRequestsAndReportErrors")

	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	pullRequests, resp, _ := client.PullRequests.List(ctx, owner, name, opt)

	for _, pr := range pullRequests {
		log.Infof(ctx, "debug: pr is %v, %v\n", pr.Title, pr.GetTitle())
	}

	log.Infof(ctx, "debug: resp is %v\n", resp)
}

func GetLabelsByIssue(ctx context.Context, issueSvc *github.IssuesService, owner string, name string, issue int) []*github.Label {
	currentLabels, _, err := issueSvc.ListLabelsByIssue(ctx, owner, name, issue, nil)
	if err != nil {
		log.Infof(ctx, "info: could not get labels by the issue")
		log.Infof(ctx, "debug: %v\n", err)
		return nil
	}
	log.Infof(ctx, "debug: the current labels: %v\n", currentLabels)
	return currentLabels
}

func PullRequestComment(webhook *Webhook, ctx context.Context) {

	log.Infof(ctx, "debug: run PullRequestComment")
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

type Repository struct {
	Owner string
	Name  string
}

type File struct {
	RawReviewers []interface{} `json:"reviewers"`
}

func getReviewersFromConf(ctx context.Context, me string) []string {
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

func getPullRequest(ctx context.Context, client *github.Client, repository Repository) {
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "closed",
	}

	var apiResponse []*github.PullRequest

	for {
		pullRequests, resp, err := client.PullRequests.List(ctx, repository.Owner, repository.Name, opt)
		if err != nil {
			// return nil, err
			return
		}
		apiResponse = append(apiResponse, pullRequests...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

// func DeleteReviewers(ctx context.Context, event *eventData) error {
// 	return this.githubClient.DeleteReviewers(ctx, event)
// }
