package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/nlopes/slack"
	"github.com/snabb/isoweek"
)

func main() {
	authToken := os.Getenv("PAGERDUTY_AUTH_TOKEN")
	if authToken == "" {
		log.Println("PAGERDUTY_AUTH_TOKEN must be set")
		os.Exit(1)
	}
	slackChannel := os.Getenv("SLACK_CHANNEL")
	if slackChannel == "" {
		log.Println("SLACK_CHANNEL must be set")
		os.Exit(1)
	}
	scheduleID := os.Getenv("PAGERDUTY_SCHEDULE_ID")
	if scheduleID == "" {
		log.Println("PAGERDUTY_SCHEDULE_ID must be set")
		os.Exit(1)
	}
	slackToken := os.Getenv("SLACK_AUTH_TOKEN")
	if slackToken == "" {
		log.Println("SLACK_AUTH_TOKEN must be set")
		os.Exit(1)
	}

	client := pagerduty.NewClient(authToken)
	slackClient := slack.New(slackToken)
	now := time.Now()
	week, year := now.ISOWeek()
	weekStart := isoweek.StartTime(week, year, time.Local).Add(time.Hour * 10)
	then := time.Now().AddDate(0, 0, 7)
	week, year = then.ISOWeek()
	weekEnd := isoweek.StartTime(week, year, time.Local)
	onCalls, err := client.ListOnCallUsers(scheduleID, pagerduty.ListOnCallUsersOptions{
		Since: weekStart.Format(time.RFC3339),
		Until: weekEnd.Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("Error retrieving on-call users: %s", err.Error())
		os.Exit(1)
	}
	users := make([]string, 0, len(onCalls))
	for _, onCall := range onCalls {
		users = append(users, onCall.Name)
	}
	var userList string
	if len(users) == 0 {
		userList = "nobody"
	} else if len(users) == 1 {
		userList = users[0]
	} else if len(users) == 2 {
		userList = users[0] + " and " + users[1]
	} else if len(users) > 2 {
		users[len(users)-1] = "and " + users[len(users)-1]
		userList = strings.Join(users, ", ")
	}
	verb := "is"
	if len(users) > 1 {
		verb = "are"
	}
	body := fmt.Sprintln("Good Morning! :coffee: Just a reminder,", userList, verb, "on bug duty :bug: this week.")
	_, _, err = slackClient.PostMessage(slackChannel, body, slack.PostMessageParameters{
		Username:    "Bug On-Duty Bot",
		AsUser:      false,
		Parse:       "full",
		IconURL:     "https://storage.googleapis.com/hashibot-public-assets/ladybug-yellow.png",
		LinkNames:   1,
		UnfurlLinks: true,
		UnfurlMedia: true,
	})
	if err != nil {
		log.Printf("Error posting in Slack: %s", err)
		os.Exit(1)
	}
	log.Println("Posted reminder in Slack.")
}
