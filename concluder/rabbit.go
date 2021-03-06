package concluder

import (
	"encoding/json"
	"fmt"
	"os"

	client "github.com/Gimulator/client-go"
	"github.com/streadway/amqp"
)

type Rabbit struct {
	uri       string
	queueName string
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func NewRabbit() (*Rabbit, error) {
	r := &Rabbit{}

	if err := r.env(); err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(r.uri)
	if err != nil {
		return nil, err
	}
	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	r.ch = ch

	return r, nil
}

func (r *Rabbit) env() error {
	r.uri = os.Getenv("LOGGER_RABBIT_URI")
	if r.uri == "" {
		return fmt.Errorf("set th 'LOGGER_RABBIT_URI' environment variable for sending result to RabbitMQ")
	}
	r.queueName = os.Getenv("LOGGER_RABBIT_QUEUE")
	if r.queueName == "" {
		return fmt.Errorf("set th 'LOGGER_RABBIT_QUEUE' environment variable for sending result to RabbitMQ")
	}
	return nil
}

func (r *Rabbit) Send(obj client.Object) error {
	defer r.conn.Close()
	defer r.ch.Close()

	queue, err := r.ch.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return err
	}

	data, err := json.Marshal(obj.Value)
	if err != nil {
		return err
	}
	body := string(data)

	if err := r.ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/x-yaml",
			Body:        []byte(body),
		},
	); err != nil {
		return err
	}

	return nil
}
