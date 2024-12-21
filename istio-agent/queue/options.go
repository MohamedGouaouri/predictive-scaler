package queue

import (
	"context"

	"github.com/streadway/amqp"
)

type QueuePublisher struct {
	queue        amqp.Queue
	queueChannel *amqp.Channel
}

type QueueConsumer struct {
	queue        amqp.Queue
	queueChannel *amqp.Channel
	tag          string
	ctx          *context.Context
}

type QueueCommand struct {
	Command string `json:"command"`
	Args    ScaleCommandArgs
}

type ScaleCommandArgs = map[string]int32
