package main

import (
	"context"
	"log"

	"github.com/lindstorm76/code_executor/pkg/queue"
)

func main() {
	q := queue.NewQueue("localhost:6379", "submissions")

	defer q.Close()

	log.Println("executor started, waiting for jobs...")

	for {
		ctx := context.Background()

		job, err := q.Dequeue(ctx)

		if err != nil {
			log.Printf("failed to dequeue: %v", err)
			
			continue
		}

		if job == nil {
			continue
		}

		log.Printf("dequeued submission %s\ncode: %s\nlanguage: %s", job.SubmissionId, job.Code, job.Language)
	}
}