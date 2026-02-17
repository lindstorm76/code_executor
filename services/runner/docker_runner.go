package runner

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerRunner struct {
	client *client.Client
}

func NewDockerRunner () (*DockerRunner, error) {
	client, err := client.New(client.FromEnv, client.WithHost("unix://" + os.Getenv("HOME") + "/.docker/run/docker.sock"))

	if err != nil {
		return nil,  err
	}

	return &DockerRunner{
		client: client,
	}, nil
}

func (d *DockerRunner) getImageName(language string) string {
	switch language {
	// On the fly.
	case "node.js": return "node:20-alpine";
	case "python":     return "python:3.11-alpine";
	case "python2":    return "python:2.7-alpine";
	case "ruby":       return "ruby:3.2-alpine";
	case "php":        return "php:8.2-alpine";
	case "perl":       return "perl:5.38-slim";
	case "lua":        return "nickblah/lua:5.4-alpine";
	default: return ""
	}
}

func (d *DockerRunner) getExecutionCommand(code, language string) ([]string, error) {
	switch language {
	case "node.js":
		return []string{"node", "-e", code}, nil
	case "python": 
		return []string{"python3", "-c", code}, nil
	case "python2":
		return []string{"python2", "-c", code}, nil
	case "ruby":
		return []string{"ruby", "-e", code}, nil
	case "php":
		return []string{"php", "-r", code}, nil
	case "perl":
		return []string{"perl", "-e", code}, nil
	case "lua":
		return []string{"lua", "-e", code}, nil
	default: return nil, fmt.Errorf("language %s not supported", language)
	}
}

func (d *DockerRunner) ExecuteCode(ctx context.Context, code, language string) (stdout, stderr string, exitCode int, err error) {
	// Get image based on language to run the code.
	image := d.getImageName(language)

	cmd, err := d.getExecutionCommand(code, language)

	if err != nil {
		return "", "", -1, fmt.Errorf("failed to get execution command: %v", err)
	}

	log.Printf("pulling docker image %s to run the code on...", image)

	reader, err := d.client.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to pull image: %v", err)
	}
	defer reader.Close()

	// Wait for the pull to complete
	io.Copy(io.Discard, reader)

	// Create container with resource limits.
	response, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: image,
			Cmd: cmd,
			AttachStdout: true,
			AttachStderr: true,
			Tty: false,
			NetworkDisabled: true,
		},
		HostConfig: &container.HostConfig{
			Resources: container.Resources{
				Memory: 512 * 512 * 1024,
				NanoCPUs: 1000000000,
				PidsLimit: func() *int64 { v := int64(50); return &v }(), // Prevent fork bombs.
			},
			ReadonlyRootfs: false,
		},
		NetworkingConfig: nil,
		Platform: nil,
		Name: "",
	})

	if err != nil {
		return "", "", -1, fmt.Errorf("container create failed: %v", err)
	}

	containerId := response.ID

	defer d.client.ContainerRemove(context.Background(), containerId, client.ContainerRemoveOptions{Force: true})

	if _, err := d.client.ContainerStart(ctx, containerId, client.ContainerStartOptions{}); err != nil {
		return "", "", -1, fmt.Errorf("container start failed: %v", err)
	}

	containerWait := d.client.ContainerWait(ctx, containerId, client.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning,
	})

	timeoutCtx, cancel := context.WithTimeout(ctx, 60 * time.Second)
	
	defer cancel()

	var statusCode int64

	select {
	case err := <-containerWait.Error:
		if err != nil {
				return "", "", -1, fmt.Errorf("container wait failed: %v", err)
		}
	case status := <-containerWait.Result:
		statusCode = status.StatusCode
	case <-timeoutCtx.Done():
		// Timeout - kill container
		d.client.ContainerKill(context.Background(), containerId, client.ContainerKillOptions{
			Signal: "SIGKILL",
		})
		return "", "", -1, fmt.Errorf("execution timeout")
	}

	// Get logs from container.
	output, err := d.client.ContainerLogs(context.Background(), containerId, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		return "", "", -1, fmt.Errorf("failed to get container logs: %v", err)
	}

	defer output.Close()

	logs, err := io.ReadAll(output)

	if err != nil {
		return "", "", -1, fmt.Errorf("failed to read logs: %v", err)
	}

	logStr := string(logs)

	if (len(logStr)) > 8 {
		logStr = logStr[8:]
	}

	return logStr, "", int(statusCode), nil
}

func (d *DockerRunner) Close() error {
    return d.client.Close()
}