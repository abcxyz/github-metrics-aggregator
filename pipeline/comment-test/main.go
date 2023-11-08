package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/google/go-github/v56/github"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		// beam/log is required in order for log severity to show up properly in
		// Dataflow. See https://github.com/abcxyz/github-metrics-aggregator/pull/171
		// for more context.
		log.Errorf(ctx, "realMain failed: %v", err)
		os.Exit(1)
	}
}

// realMain executes the PR Comment Pipeline.
func realMain(ctx context.Context) error {
	pat := os.Getenv("github-pat")
	client := github.NewClient(nil).WithAuthToken(pat)
	messageText := "This is a test of the api. Please ignore!"
	comment := &github.IssueComment{
		Body: &messageText,
	}
	org := "abcxyz"
	repo := "github-metrics-aggregator"
	prNumber := 165
	numberResults := 2
	// Sleep from [0, 1000 * numberResults) ms to have an average 1.0 QPS.
	sleepMs := rand.Intn(1_000 * numberResults)
	fmt.Printf("Sleeping for %v ms.\n", sleepMs)
	time.Sleep(time.Millisecond * time.Duration(sleepMs))
	fmt.Printf("OK, I'm awake!\n")

	createdComment, response, err := client.Issues.CreateComment(ctx, org, repo, prNumber, comment)
	if err != nil {
		// Treat rate limit errors as any other, let future jobs retry.
		if errors.As(err, new(*github.RateLimitError)) {
			return fmt.Errorf("lol, got a rate limit error: %w", err)
		} else if errors.As(err, new(*github.AbuseRateLimitError)) {
			return fmt.Errorf("oof, got a secondary rate limit error: %w", err)
		} else {
			return fmt.Errorf("whelp, some random error: %w", err)
		}
	}
	println(response)
	println(createdComment)
	deleteResponse, err := client.Issues.DeleteComment(ctx, org, repo, *createdComment.ID)
	if err != nil {
		// Treat rate limit errors as any other, let future jobs retry.
		if errors.As(err, new(*github.RateLimitError)) {
			return fmt.Errorf("lol, got a rate limit error: %w", err)
		} else if errors.As(err, new(*github.AbuseRateLimitError)) {
			return fmt.Errorf("oof, got a secondary rate limit error: %w", err)
		} else {
			return fmt.Errorf("whelp, some random error: %w", err)
		}
	}
	print(deleteResponse)

	return nil
}
