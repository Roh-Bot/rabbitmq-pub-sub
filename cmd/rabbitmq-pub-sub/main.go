package main

import (
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/config"
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/db"
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/messagebrokers"
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/global"
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/loggers"
	"log"
	"sync"
	"time"
)

func main() {
	// Parsing global flags
	global.LoadGlobalFlags()

	// Loading configuration
	if err := config.LoadConfiguration(); err != nil {
		log.Fatal(err)
		return
	}

	// Setting up logger
	if err := loggers.ZapNew(); err != nil {
		log.Fatal(err)
		return
	}

	// Connecting to Postgres
	if err := db.PostgresNewPool(); err != nil {
		log.Fatal(err)
		return
	}

	// Connecting to rabbit mq
	if err := messagebrokers.RabbitMQConnect(); err != nil {
		log.Fatal(err)
		return
	}

	n := 5
	// Making a wait group to handle graceful shutdown
	wg := new(sync.WaitGroup)

	wg.Add(n)
	// Firing up 10 consumers
	for i := 0; i < n; i++ {
		go messagebrokers.RabbitMQ().ReceiveMessages(wg)
		time.Sleep(time.Millisecond * 200)
	}

	message := map[string]any{"Rabbit": "Kafka"}
	messagebrokers.RabbitMQ().SendMessages(message)

	<-global.CancellationContext().Done()
	wg.Wait()

	messagebrokers.RabbitMQ().DeleteAllQueues()
	messagebrokers.RabbitMQ().Shutdown()

	loggers.Zap.Sync()
}
