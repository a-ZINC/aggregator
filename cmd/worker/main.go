package main

import (
	"context"
	"fmt"
	"log"

	// "github.com/a-ZINC/aggregator/config"
	// "github.com/a-ZINC/aggregator/internal/db"
	"github.com/a-ZINC/aggregator/internal/domain"
	// "github.com/a-ZINC/aggregator/internal/fetcher"
	"github.com/a-ZINC/aggregator/internal/notifier"
	"github.com/a-ZINC/aggregator/internal/store"
)

func main() {
	// cfg := config.Load()
	// dbConfig := config.BuildDbConfig(cfg)
	// sql, err := db.NewConnectionPool(dbConfig)
	// if err != nil {
	// 	log.Fatalf("failed to connect to database: %v", err)
	// }
	// jobStore := store.NewJobStore(sql)
	// defer jobStore.Close()
	// appStore := store.NewStore(jobStore)
	
	// remotiveFetcher := fetcher.NewRemotiveFetcher()
	// jobs, err := remotiveFetcher.Fetch(context.Background())
	// if err != nil {
	// 	log.Printf("fetch failed Remotive: %v", err)
	// }

	// remoteOkFetcher := fetcher.NewRemoteOkFetcher()
	// remoteOkJobs, err := remoteOkFetcher.Fetch(context.Background())
	// if err != nil {
	// 	log.Printf("fetch failed RemoteOk: %v", err)
	// }
	// log.Printf("Fetched %d jobs from Remotive and %d jobs from RemoteOk\n", len(jobs), len(remoteOkJobs))

	// for i, j := range jobs {
	// 	if i >= 3 {
	// 		break
	// 	}
	// 	fmt.Printf("\n[%d] %s\n", i+1, j.Title)
	// 	fmt.Printf("    Company  : %s\n", j.Company)
	// 	fmt.Printf("    Location : %s\n", j.Location)
	// 	fmt.Printf("    URLHash  : %s\n", j.UrlHash)
	// 	fmt.Printf("    Tags     : %v\n", j.Tags[:3])
	// 	fmt.Printf("    PostedAt : %v\n", j.PostedAt)
	// 	fmt.Printf("    Source   : %s\n", j.Source)
	// }

	// for i, j := range remoteOkJobs {
	// 	if i >= 3 {
	// 		break
	// 	}
	// 	fmt.Printf("\n[%d] %s\n", i+1, j.Title)
	// 	fmt.Printf("    Company  : %s\n", j.Company)
	// 	fmt.Printf("    Location : %s\n", j.Location)
	// 	fmt.Printf("    URLHash  : %s\n", j.UrlHash)
	// 	fmt.Printf("    Tags     : %v\n", j.Tags[:3])
	// 	fmt.Printf("    PostedAt : %v\n", j.PostedAt)
	// 	fmt.Printf("    Source   : %s\n", j.Source)
	// }
	// err = buildJob(appStore, &remoteOkJobs[0])
	// if err != nil {
	// 	log.Printf("failed to build job: %v", err)
	// }
	sendMessage()


}

func sendMessage() {
	notifier := notifier.NewTelegramNotifier()
	err := notifier.Notify(context.Background(), &domain.Job{
		Title:   "Senior Software Engineer",
		Company: "TechCorp",
		Source:  "Remotive",
		ApplyUrl: "https://remotive.io/remote-jobs/software-dev/senior-software-engineer-123456",
	}, "High Signal Match")
	if err != nil {
		log.Printf("failed to send notification: %v", err)
	}
}

func buildJob(appStore *store.Store, job *domain.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	exist, err := appStore.JobStore.Exists(context.Background(), job.UrlHash)
	if err != nil {
		return fmt.Errorf("failed to check if job exists: %v", err)
	} else if exist {
		return fmt.Errorf("job already exists: %s", job.UrlHash)
	}
	
	if err := appStore.JobStore.Save(context.Background(), job); err != nil {
		return fmt.Errorf("failed to insert job: %v", err)
	}
	
	return nil
}