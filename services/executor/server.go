package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"

	pb "github.com/lindstorm76/code_executor/api/pb/api/proto"
	"github.com/lindstorm76/code_executor/services/runner"
	"google.golang.org/grpc"
)

type executorServer struct {
	pb.UnimplementedExecutorServiceServer

	dockerRunner *runner.DockerRunner
}

func (s *executorServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	log.Printf("executing submission %s inside docker container", req.SubbmissionId)

	// Execute the command.
	stdout, stderr, exitCode, err  := s.dockerRunner.ExecuteCode(ctx, req.Code, req.Language)

	status := "SUCCESS"

	if err != nil {
		if err.Error() == "execution timeout" {
			status = "TIME_LIMIT_EXCEEDED"
		} else {
			status = "RUNTIME_ERROR"
			stderr = err.Error()
		}
	} else if exitCode != 0 {
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
	dockerRunner, err := runner.NewDockerRunner()

	if err != nil {
		log.Fatalf("failed to initiate docker runner: %v", err)
	}

	defer dockerRunner.Close()

	listener, err := net.Listen("tcp", ":3002")

	if err != nil {
		log.Fatalf("failed to listen to :3002")
	}

	grpcServer := grpc.NewServer()

	pb.RegisterExecutorServiceServer(grpcServer, &executorServer{
		dockerRunner: dockerRunner,
	})

	log.Println("executor server (docker) listening on :3002")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}