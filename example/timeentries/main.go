package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gogolok/go-harvest/harvest"
)

func main() {
	accessToken := os.Getenv("TOKEN")
	if len(accessToken) < 1 {
		log.Fatalf("You MUST set environment variable TOKEN")
	}
	accountId := os.Getenv("ACCOUNT_ID")
	if len(accountId) < 1 {
		log.Fatalf("You MUST set environment variable ACCOUNT_ID")
	}

	from := time.Now().AddDate(0, 0, -14).Format("20060102")
	to := time.Now().Format("20060102")
	ctx := context.Background()
	opts := &harvest.TimeEntriesListOptions{
		From:        from,
		To:          to,
		ListOptions: harvest.ListOptions{Page: 1, PerPage: 100},
	}
	client := harvest.NewClient(accessToken, accountId)

	timeEntries, r, err := client.TimeEntries.List(ctx, opts)
	if err != nil {
		log.Fatalf("Error fetching time entries: %v+\n", err)
	}
	log.Printf("response = %+v\n", r)
	log.Printf("No. of entries: %+v\n", len(timeEntries))
}
