package rabbitmq

import (
	"log"
	"search-courses-api/src/config/envs"
	"sync"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	QueueName  string
}

var instance *RabbitMQ
var once sync.Once

func NewRabbitMQ() *RabbitMQ {
	once.Do(func() {
		env := envs.LoadEnvs()
		amqpURL := env.Get("RABBITMQ_URL")
		queueName := env.Get("RABBITMQ_QUEUE_NAME")
		if queueName == "" {
			queueName = "course_updates"
		}

		conn, err := amqp.Dial(amqpURL)
		if err != nil {
			log.Fatalf("Error al conectar con RabbitMQ: %v", err)
		}

		ch, err := conn.Channel()
		if err != nil {
			log.Fatalf("Error al abrir un canal en RabbitMQ: %v", err)
		}

		_, err = ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			log.Fatalf("Error al declarar la cola en RabbitMQ: %v", err)
		}

		instance = &RabbitMQ{
			connection: conn,
			channel:    ch,
			QueueName:  queueName,
		}

		log.Println("Conexión a RabbitMQ establecida y cola declarada")
	})

	return instance
}

func (r *RabbitMQ) ConsumeMessages(handler func(message string)) {
	msgs, err := r.channel.Consume(
		r.QueueName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Fatalf("Error al consumir mensajes de RabbitMQ: %v", err)
	}

	go func() {
		for msg := range msgs {
			message := string(msg.Body)
			log.Printf("Mensaje recibido de RabbitMQ: %s", message)
			handler(message)
		}
	}()
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
	log.Println("Conexión a RabbitMQ cerrada")
}
