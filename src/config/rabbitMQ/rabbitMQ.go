package rabbitMQ

import (
	"log"
	"search-courses-api/src/config/envs"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	connection     *amqp.Connection
	channel        *amqp.Channel
	QueueName      string
	amqpURL        string
	mu             sync.RWMutex
	messageHandler func(message string)
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

		instance = &RabbitMQ{
			QueueName: queueName,
			amqpURL:   amqpURL,
		}

		go instance.connectWithRetry()
	})

	return instance
}

func (r *RabbitMQ) connectWithRetry() {
	for {
		conn, err := amqp.Dial(r.amqpURL)
		if err != nil {
			log.Printf("Error al conectar con RabbitMQ: %v. Reintentando en 5 segundos...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			log.Printf("Error al abrir un canal en RabbitMQ: %v. Reintentando en 5 segundos...", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		_, err = ch.QueueDeclare(
			r.QueueName, // name
			true,        // durable
			false,       // delete when unused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
		)
		if err != nil {
			log.Printf("Error al declarar la cola en RabbitMQ: %v. Reintentando en 5 segundos...", err)
			ch.Close()
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		r.mu.Lock()
		r.connection = conn
		r.channel = ch
		r.mu.Unlock()

		log.Println("Conexión a RabbitMQ establecida y cola declarada.")

		// Iniciar el consumo de mensajes si el handler está establecido
		r.mu.RLock()
		handlerSet := r.messageHandler != nil
		r.mu.RUnlock()

		if handlerSet {
			r.startConsuming()
		}

		// Manejar la reconexión si la conexión se pierde
		closeChan := make(chan *amqp.Error)
		r.connection.NotifyClose(closeChan)

		err = <-closeChan
		if err != nil {
			log.Printf("Conexión a RabbitMQ cerrada: %v. Reintentando conexión...", err)
		}

		r.mu.Lock()
		r.channel = nil
		r.connection = nil
		r.mu.Unlock()

		// Esperar antes de reintentar
		time.Sleep(5 * time.Second)
	}
}

func (r *RabbitMQ) startConsuming() {
	r.mu.RLock()
	ch := r.channel
	handler := r.messageHandler
	r.mu.RUnlock()

	if ch == nil || handler == nil {
		log.Println("Canal de RabbitMQ o handler no están listos. No se puede consumir mensajes.")
		return
	}

	msgs, err := ch.Consume(
		r.QueueName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Printf("Error al consumir mensajes de RabbitMQ: %v", err)
		return
	}

	go func() {
		for msg := range msgs {
			message := string(msg.Body)
			log.Printf("Mensaje recibido de RabbitMQ: %s", message)
			handler(message)
		}
	}()
}

func (r *RabbitMQ) ConsumeMessages(handler func(message string)) {
	r.mu.Lock()
	r.messageHandler = handler
	r.mu.Unlock()

	// Si el canal está listo, comenzar a consumir
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch != nil {
		r.startConsuming()
	}
}

func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
	log.Println("Conexión a RabbitMQ cerrada")
}
