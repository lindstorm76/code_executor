package main

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/lindstorm76/code_executor/api/pb/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSubmission(t *testing.T) {
	conn, err := grpc.NewClient("localhost:3001", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
			t.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewSubmissionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	// Try submitting a code.
	response, err := client.Submit(ctx, &pb.SubmitRequest{
		Code: "console.log('Hello from Node.js!')",
		Language: "node.js",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Try submitting another code.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: "print('Hello from Python!')",
		Language: "python",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Try getting the status of previously submitted code.
	status, err := client.GetStatus(ctx, &pb.GetStatusRequest{
		SubmissionId: response.SubmissionId,
	})

	if err != nil {
		t.Fatalf("could not get status of %s: %v", response.SubmissionId, err)
	}

	log.Printf("status of %s: %s", response.SubmissionId, status)
}