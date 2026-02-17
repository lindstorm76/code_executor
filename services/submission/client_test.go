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

	// JavaScript.
	response, err := client.Submit(ctx, &pb.SubmitRequest{
		Code: `console.log("Hello from JavaScript!")`,
		Language: "node.js",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Python.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `print("Hello from Python!")`,
		Language: "python",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Python2.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `print "Hello from Python2!"`,
		Language: "python2",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Ruby.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `puts "Hello from Ruby!"`,
		Language: "ruby",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// PHP.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `echo "Hello from PHP!";`,
		Language: "php",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Perl.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `print "Hello from Perl!\n";`,
		Language: "perl",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// Lua.
	response, err = client.Submit(ctx, &pb.SubmitRequest{
		Code: `print("Hello from Lua!")`,
		Language: "lua",
	})

	if err != nil {
		t.Fatalf("could not submit: %v", err)
	}

	log.Printf("submission id:  %s", response.SubmissionId)

	// // Try getting the status of previously submitted code.
	// status, err := client.GetStatus(ctx, &pb.GetStatusRequest{
	// 	SubmissionId: response.SubmissionId,
	// })

	// if err != nil {
	// 	t.Fatalf("could not get status of %s: %v", response.SubmissionId, err)
	// }

	// log.Printf("status of %s: %s", response.SubmissionId, status)
}