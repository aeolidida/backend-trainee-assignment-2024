package rabbitmq

import (
	"backend-trainee-assignment-2024/internal/config"
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQQueue struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queueName  string
}

func NewRabbitMQQueue(cfg config.RabbitMQ) (*RabbitMQQueue, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Address, cfg.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		cfg.Name,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQQueue{
		connection: conn,
		channel:    ch,
		queueName:  cfg.Name,
	}, nil
}

func (q *RabbitMQQueue) Publish(message []byte, queue string) error {
	return q.channel.PublishWithContext(
		context.Background(),
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
}

func (q *RabbitMQQueue) Consume(queue string) (<-chan []byte, error) {
	msgs, err := q.channel.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	messages := make(chan []byte, 100)
	go func() {
		for msg := range msgs {
			messages <- msg.Body
		}
	}()

	return messages, nil
}

func (q *RabbitMQQueue) Close() error {
	err := q.channel.Close()
	if err != nil {
		return err
	}

	err = q.connection.Close()
	if err != nil {
		return err
	}

	return nil
}
