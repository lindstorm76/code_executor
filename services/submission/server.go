package main

import (
	"context"
	"log"
	"net"

	"github.com/google/uuid"
	pb "github.com/lindstorm76/code_executor/api/pb/api/proto"
	"google.golang.org/grpc"
)

type submissionServer struct {
	pb.UnimplementedSubmissionServiceServer

	submissions map[string]*pb.GetStatusResponse
}

func (s *submissionServer) Submit(ctx context.Context, req *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	submissionId := uuid.New().String()

	s.submissions[submissionId] = &pb.GetStatusResponse{
		Status: "PENDING",
	}

	log.Printf("submission recieved: %s (language: %s)", submissionId, req.Language);

	return &pb.SubmitResponse{
		SubmissionId: submissionId,
	}, nil
}

func (s *submissionServer) GetStatus(ctx context.Context, req*pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	status, exists := s.submissions[req.SubmissionId]

	if !exists {
		return &pb.GetStatusResponse{
			Status: "NOT_FOUND",
		}, nil
	}

	return status, nil
}

func main() {
	listener, err := net.Listen("tcp", ":3001")

	if err != nil {
		log.Fatalf("failed to listen to :3001")
	}

	grpcServer := grpc.NewServer()

	pb.RegisterSubmissionServiceServer(grpcServer, &submissionServer{
		submissions: make(map[string]*pb.GetStatusResponse),
	})

	log.Println("submission server listening on :3001")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}