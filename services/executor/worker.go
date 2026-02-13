package main

import (
	"context"
	"log"

	pb "github.com/lindstorm76/code_executor/api/pb/api/proto"
	"github.com/lindstorm76/code_executor/pkg/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect the worker to execution server.
	conn, err := grpc.NewClient("localhost:3002", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect to execution server: %v", err)
	}

	defer conn.Close()

	executorClient := pb.NewExecutorServiceClient(conn)

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

		log.Printf("dequeued submission %s, starting execution...", job.SubmissionId)
		
		result, err := executorClient.Execute(ctx, &pb.ExecuteRequest{
			SubbmissionId: job.SubmissionId,
			Code: job.Code,
			Language: job.Language,
		})

		if err != nil {
			log.Printf("execution error: %v", err)

			continue
		}

		log.Printf("execution completed %s", result.SubmissionId)
		log.Printf("output: %s", result.Stdout)
	}
}