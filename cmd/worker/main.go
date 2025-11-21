package main

import (
	"kosis/internal/config"
	"kosis/internal/db"
	"kosis/internal/tasks"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Worker connected to database.")

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{})
	fetchReportsTask, err := tasks.NewFetchReportsTask(nil, nil)
	if err != nil {
		log.Fatalf("Failed to create fetch reports task: %v", err)
	}

	// every 5 minutes
	entryID, err := scheduler.Register("*/5 * * * *", fetchReportsTask, asynq.Queue("default"))
	if err != nil {
		log.Fatalf("Failed to register periodic task: %v", err)
	}
	log.Printf("Registered periodic task: %s (EntryID: %s)", fetchReportsTask.Type(), entryID)

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Specify different queues or priorities if needed
			Queues: map[string]int{
				"default": 3,
			},
			Concurrency: 10, // Max 10 concurrent jobs
		},
	)

	taskProcessor := tasks.NewTaskProcessor(db, cfg)

	mux := asynq.NewServeMux()
	mux.HandleFunc(
		tasks.TypeTaskFetchReports,
		taskProcessor.HandleFetchReportsTask,
	)

	mux.HandleFunc(
		tasks.TypeTaskFetchCompanies,
		taskProcessor.HandleFetchCompaniesTask,
	)

	// if _, err := asynqClient.Enqueue(fetchReportsTask); err != nil {
	// 	log.Fatalf("Failed to enqueue fetch reports task: %v", err)
	// }

	// if _, err := asynqClient.Enqueue(fetchCompaniesTask); err != nil {
	// 	log.Fatalf("Failed to enqueue fetch companies task: %v", err)
	// }

	go func() {
		log.Println("Starting Asynq scheduler...")
		if err := scheduler.Run(); err != nil {
			log.Fatalf("Could not run Asynq scheduler: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Asynq worker server...")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Could not run Asynq worker server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutdown signal received, shutting down gracefully...")

	scheduler.Shutdown()
	log.Println("Asynq scheduler shut down.")

	srv.Shutdown()
	log.Println("Asynq worker server shut down.")

	asynqClient.Close()
	log.Println("Asynq client closed.")

	log.Println("Worker process shut down complete.")
}
