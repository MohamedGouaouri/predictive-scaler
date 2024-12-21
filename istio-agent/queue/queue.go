package queue

import (
	"fmt"
	"log"

	"github.com/MohammedGouaouri/get-pod-metrics/constants"
	"github.com/MohammedGouaouri/get-pod-metrics/utils"
	"github.com/streadway/amqp"
)

func NewPublisher(name string) *QueuePublisher {
	publisher := &QueuePublisher{}
	conn, err := amqp.Dial(constants.QUEUE_URL)
	utils.FailOnError(err, "Failed to connect to RabbitMQ")

	// Create a channel
	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a publish channel")
	publisher.queueChannel = ch

	// Declare a queue
	q, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	utils.FailOnError(err, "Failed to declare a queue")

	publisher.queue = q
	return publisher

}

func NewCosnumer(name string, tag string) *QueueConsumer {
	consumer := &QueueConsumer{}
	conn, err := amqp.Dial(constants.QUEUE_URL)
	utils.FailOnError(err, "Failed to connect to RabbitMQ")

	// Create a channel
	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a publish channel")
	consumer.queueChannel = ch

	// Declare a queue
	q, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	utils.FailOnError(err, "Failed to declare a queue")

	consumer.queue = q
	consumer.tag = tag
	return consumer

}

func (publisher *QueuePublisher) Publish(message string) {
	err := publisher.queueChannel.Publish(
		"",                   // exchange
		publisher.queue.Name, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	utils.FailOnError(err, "Failed to publish a message")
	fmt.Printf(" [x] Sent %s\n", message)
}

func (consumer *QueueConsumer) Consume(commandChan chan string) {
	msgs, err := consumer.queueChannel.Consume(
		consumer.queue.Name, // queue
		consumer.tag,        // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			commandChan <- string(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
