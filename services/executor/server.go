package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	pb "github.com/lindstorm76/code_executor/api/pb/api/proto"
	"google.golang.org/grpc"
)

type executorServer struct {
	pb.UnimplementedExecutorServiceServer
}

func (s *executorServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	log.Printf("executing submission %s", req.SubbmissionId)

	// Get command to execute.
	cmd, err := getExecutionCommand(req.Code, req.Language)

	if err != nil {
		log.Printf("failed to get execution command for submission %s", req.SubbmissionId)
		
		return nil, err
	}

	// Set timeout for the execution, allowing up to 60 seconds.
	execCtx, cancel := context.WithTimeout(ctx, 60 * time.Second)

	defer cancel()

	// Execute the command.
	stdout, stderr, exitCode, err := executeCommand(ctx, cmd)
	status := "SUCCESS"

	if execCtx.Err() == context.DeadlineExceeded {
		status = "TIME_LIMIT_EXCEEDED"
	} else if exitCode == 0 {
		status = "RUNTIME_ERROR"
	}

	return &pb.ExecuteResponse{
		SubmissionId: req.SubbmissionId,
		Stdout: stdout,
		Stderr: stderr,
		ExitCode: int32(exitCode),
		Status: status,
	}, nil
}

func getExecutionCommand(code, language string) (*exec.Cmd, error) {
	switch language {
	case "python": 
		return exec.Command("python3", "-c", code), nil
	case "node.js":
		return exec.Command("node", "-e", code), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

func executeCommand(ctx context.Context, cmd *exec.Cmd) (stdout, stderr string, exitCode int, err error) {
	stdoutBytes, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(stdoutBytes), string(exitErr.Stderr), exitErr.ExitCode(), nil
		}

		return "", "", -1, err
	}

	return string(stdoutBytes), "", 0, nil
}

func main() {
	listener, err := net.Listen("tcp", ":3002")

	if err != nil {
		log.Fatalf("failed to listen to :3002")
	}

	grpcServer := grpc.NewServer()

	pb.RegisterExecutorServiceServer(grpcServer, &executorServer{})

	log.Println("executor server listening on :3002")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}