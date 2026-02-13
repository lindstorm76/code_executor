package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	SubmissionId string `json:"submission_id"`
	Code string `json:"code"`
	Language string `json:"language"`
}

type Queue struct {
	client *redis.Client
	queueName string
}

func NewQueue(addr string, queueName string) *Queue {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Queue{
		client: client,
		queueName: queueName,
	}
}

func (q *Queue) Enqueue(ctx context.Context, job *Job) error {
	data, err := json.Marshal(job)

	if err != nil {
		return err
	}

	return q.client.LPush(ctx, q.queueName, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context) (*Job, error) {
	result, err := q.client.BRPop(ctx, 5 * time.Second).Result()

	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, nil
	}

	var job Job

	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (q *Queue) Close() error {
	return q.client.Close()
}