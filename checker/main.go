package main

import (
	"encoding/json"
	mdb "github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-checker/checker/runner"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func checker(period int, done chan bool) {
	run := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Duration(period) * time.Second)
			run <- true
		}
	}()
	for {
		select {
		case <-done:
			log.Println("Done!")
			return
		case <-run:
			err := runner.RunChecker()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	mdb.InitMongo()
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	done := make(chan bool, 1)
	go func() {
		for d := range msgs {
			var m Message
			log.Printf("Received a message: %s", d.Body)
			err := json.Unmarshal(d.Body, &m)
			if err != nil {
				log.Fatal(err)
			}
			switch m.Type {
			case "start":
				go checker(1, done)
			case "stop":
				done <- true
			}
		}
	}()
	go log.Printf("Checker startes")
	<-forever
}
