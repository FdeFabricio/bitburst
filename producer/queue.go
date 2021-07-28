package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Queue struct {
	connection *amqp.Connection
	name       string
}

type Message struct {
	ID int
}

func (q *Queue) connect() error {
	if q.connection == nil {
		conn, err := amqp.Dial(os.Getenv("AMQP_URI"))
		if err != nil {
			return extendError(err, "queue: failed to connect")
		}
		q.connection = conn
	}

	q.name = os.Getenv("QUEUE_NAME")

	ch, err := q.connection.Channel()
	if err != nil {
		return extendError(err, "queue: failed to open a channel")
	}

	_, err = ch.QueueDeclare(
		q.name,
		true,  // durable true: queues are persisted
		false, // autoDelete false: do not delete when unused
		false, // exclusive false: can be used by multiple connections
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return extendError(err, "queue: failed to create queue")
	}

	err = ch.Close()
	if err != nil {
		return extendError(err, "queue: failed to close channel")
	}

	return nil
}

func (q *Queue) publish(ID int) error {
	ch, err := q.connection.Channel()
	if err != nil {
		return extendError(err, "queue: failed to open channel")
	}

	message := Message{ID: ID}
	body, err := json.Marshal(message)
	if err != nil {
		return extendError(err, "queue: failed to encode message")
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         body,
	}

	err = ch.Publish("", q.name, false, false, msg)
	if err != nil {
		return extendError(err, "queue: failed to publish message")
	}

	err = ch.Close()
	if err != nil {
		return extendError(err, "queue: failed to close channel")
	}

	return nil
}
