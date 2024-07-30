package messagebrokers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/config"
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/db"
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/db/functions"
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/global"
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/loggers"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

var rabbitMQ *RabbitMq

type RabbitMq struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	p    *Publisher
}

type Publisher struct {
	messages chan map[string]any
	subs     []chan map[string]any
	mutex    *sync.RWMutex
}

func RabbitMQConnect() error {
	log.Println("Connecting to RabbitMq")

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		config.GetConfig().RabbitMQ.User,
		config.GetConfig().RabbitMQ.Password,
		config.GetConfig().RabbitMQ.Host,
		config.GetConfig().RabbitMQ.Port)

	conn, err := amqp091.Dial(url)
	if err != nil {
		loggers.Zap.Errorf("RabbitMQ Error: %s", err.Error())
		return err
	}

	if conn.IsClosed() {
		loggers.Zap.Errorf("RabbitMQ connection is closed")
		return fmt.Errorf("RabbitMQ connection is closed")
	}

	ch, err := conn.Channel()
	if err != nil {
		loggers.Zap.Errorf("RabbitMQ Error: %s", err.Error())
		return err
	}

	rabbitMQ = &RabbitMq{
		conn: conn,
		ch:   ch,
	}
	rabbitMQ.p = &Publisher{
		messages: make(chan map[string]any),
		subs:     make([]chan map[string]any, 0),
		mutex:    new(sync.RWMutex),
	}
	log.Println("Connected to RabbitMQ")
	return nil
}

func RabbitMQ() *RabbitMq {
	return rabbitMQ
}

func (r *RabbitMq) Publisher() *Publisher {
	return r.p
}

func (r *RabbitMq) SendMessages(body map[string]any) {
	if r.conn.IsClosed() || r.ch.IsClosed() {
		loggers.Zap.Errorf("RabbitMQ connection/channel is closed")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		return
	}

	var queues []string
	rows, err := db.Postgres().Read(functions.GetAllQueues)
	if err != nil {
		loggers.Zap.Errorf("PG Error: %s", err.Error())
		return
	}
	for rows.Next() {
		err := rows.Scan(&queues)
		if err != nil {
			loggers.Zap.Errorf("PG Error: %s", err.Error())
			return
		}
	}

	for _, queue := range queues {
		err := r.ch.QueueBind(
			queue,
			"",
			config.GetConfig().RabbitMQ.Exchange,
			false,
			nil,
		)
		if err != nil {
			return
		}
	}

	err = r.ch.PublishWithContext(
		ctx,
		config.GetConfig().RabbitMQ.Exchange,
		"",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        bodyBytes,
		})
	if err != nil {
		loggers.Zap.Errorf("%s : %s", err, "Failed to publish a message")
		return
	}

	log.Printf("RABBITMQ MESSAGE SENT: Sent %v\n", body)
	return
}

func (r *RabbitMq) ReceiveMessages(wg *sync.WaitGroup) {
	defer wg.Done()

	if r.conn.IsClosed() || r.ch.IsClosed() {
		log.Println("RabbitMQ connection is already closed")
		return
	}

	var queueName string
	if err := db.PGReadSingleRow(&queueName, functions.GenerateAndGetQueueName); err != nil {
		log.Println(err)
		return
	}

	defer func() {
		log.Printf("Deleting queue: %s\n", queueName)
		if err := db.Postgres().ExecNonQuery(functions.DeleteQueue, queueName); err != nil {
			loggers.Zap.Errorf("PG Error: %s", err.Error())
			return
		}
	}()

	q, err := r.ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,
	)
	if err != nil {
		return
	}
	msgs, err := r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		loggers.Zap.Errorf("RabbitMQ Error: %s", err.Error())
		return
	}

	log.Printf(" [*] Waiting for messages from queue: %s. To exit press CTRL+C", queueName)
loop:
	for {
		select {
		case <-global.CancellationContext().Done():
			log.Println("Exiting RabbitMQ receiver...")
			break loop
		case deliver, ok := <-msgs:
			if !ok {
				log.Printf("Receiver channel for queue %s has been closed. Exiting...", queueName)
				break loop
			}
			body := make(map[string]any)
			if err := json.Unmarshal(deliver.Body, &body); err != nil {
				loggers.Zap.Errorf("RabbitMQ Error: %s", err.Error())
				break
			}
			r.p.publishMessage(body)
			log.Printf("Received a message: %v", body)
		}
	}
	log.Printf("RabbitMQ receiver for queue %s stopped", queueName)
}

func (*RabbitMq) DeleteAllQueues() {
	if err := db.Postgres().ExecNonQuery(functions.TruncateQueues); err != nil {
		loggers.Zap.Errorf("PG Error: %s", err.Error())
		return
	}
}

func (r *RabbitMq) Shutdown() {
	log.Println("Closing rabbit connection...")
	if r.conn.IsClosed() {
		log.Println("RabbitMQ connection is already closed")
		return
	}
	err := r.conn.Close()
	if err != nil {
		loggers.Zap.Errorf("Failed to close connection: %s", err)
	}
	log.Println("Rabbit connection closed successfully")

	log.Println("Closing rabbit channel...")
	err = r.ch.Close()
	if err != nil {
		loggers.Zap.Errorf("Failed to close channel: %s", err)
	}
	log.Println("Rabbit channel closed successfully")
}

func (p *Publisher) publishMessage(body map[string]any) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	for _, sub := range p.subs {
		sub <- body
	}
}

func (p *Publisher) SubscribeMessages(sub chan map[string]any) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.subs = append(p.subs, sub)
}

func (p *Publisher) Shutdown() {
	log.Println("Closing all RabbitMq messaging subscribers")
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, sub := range p.subs {
		close(sub)
	}
	log.Println("Closed all RabbitMq messaging subscribers")
}
