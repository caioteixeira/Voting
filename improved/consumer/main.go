package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/streadway/amqp"
)

type vote struct {
	Target string `json:"target"`
}

type voteMessage struct {
	Target string `json:"target"`
	ID string `json:"id"`
}

type entry struct {
	ID    string `json:"id"`
	Votes int    `json:"votes"`
}

var db *pgxpool.Pool

var queue amqp.Queue

var connection *amqp.Connection

func postVote(message amqp.Delivery) {
	var newVote voteMessage

	err := json.Unmarshal(message.Body, &newVote)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal message: %v\n", err)
		return
	}

	_, err = db.Exec(context.Background(), "INSERT INTO votes(id, target) VALUES($1, $2);", newVote.ID, newVote.Target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Insert Row failed: %v\n", err)
		return
	}

	log.Printf("Processed vote: %s %s", newVote.ID, newVote.Target)
}

func main() {
	var err error
	dbconfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	db, err = pgxpool.ConnectConfig(context.Background(), dbconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}


	connection, err = amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to RabbitMQ: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open a channel: %v\n", err)
		os.Exit(1)
	}
	defer ch.Close()

	queue, err = ch.QueueDeclare(
		"votes-queue", // name
		true,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to declare a queue: %v\n", err)
		os.Exit(1)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register a consumer: %v\n", err)
		os.Exit(1)
	}

	forever := make(chan bool)

	go func() {
		for message := range msgs {
			postVote(message)
		}
	}()

	print(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
