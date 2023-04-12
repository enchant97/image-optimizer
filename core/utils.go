package core

import (
	"log"

	"github.com/enchant97/image-optimizer/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Return the minimum of two integers
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Panic on error and log the error
func PanicOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

type RabbitMQ struct {
	Conn  *amqp.Connection
	Ch    *amqp.Channel
	Queue amqp.Queue
}

func (r *RabbitMQ) Close() error {
	if err := r.Ch.Close(); err != nil {
		return err
	}
	if err := r.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) Connect(config config.AMPQConfig) error {
	conn, err := amqp.Dial(config.URI)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	q, err := ch.QueueDeclare(
		config.QueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}
	r.Conn = conn
	r.Ch = ch
	r.Queue = q
	return nil
}
