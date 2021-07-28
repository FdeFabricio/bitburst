package main

import (
	"os"

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

func (q *Queue) consume() (<-chan amqp.Delivery, *amqp.Channel, error) {
	ch, err := q.connection.Channel()
	if err != nil {
		return nil, ch, extendError(err, "queue: failed to open channel")
	}

	// with autoAck set to false we guarantee that a message will only be removed from
	// the queue after it has been correctly process, saved to PostgreSQL and then acked
	messages, err := ch.Consume(q.name, "", false, false, false, false, nil)
	if err != nil {
		return nil, ch, extendError(err, "queue: failed to start consuming queued messages")
	}

	return messages, ch, nil
}
