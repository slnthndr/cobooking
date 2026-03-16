package mq

import (
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQConn(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	// Объявляем очередь (durable=true, чтобы сообщения не пропадали при рестарте)
	_, err = ch.QueueDeclare(
		"EventsQueue", // Имя очереди
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, nil, err
	}

	log.Println("Connected to RabbitMQ successfully")
	return conn, ch, nil
}
