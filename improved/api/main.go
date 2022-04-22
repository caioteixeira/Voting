package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/streadway/amqp"
)

type vote struct {
	Target string `json:"target"`
}

type entry struct {
	ID    string `json:"id"`
	Votes int    `json:"votes"`
}

var db *pgxpool.Pool

var queue amqp.Queue

var connection *amqp.Connection

func getVoteCount(c *gin.Context) {
	var err error
	var rows pgx.Rows
	rows, err = db.Query(context.Background(), "select target, count(*) from votes group by target")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	var votes = []entry{}

	for rows.Next() {
		var newEntry entry

		if err := rows.Scan(&newEntry.ID, &newEntry.Votes); err != nil {
			fmt.Fprintf(os.Stderr, "Query Scan failed: %v\n", err)
			continue
		}
		fmt.Printf("%s has %d votes\n", newEntry.ID, newEntry.Votes)
		votes = append(votes, newEntry)
	}

	c.IndentedJSON(http.StatusOK, votes)
}

func postVote(c *gin.Context) {
	var newVote vote

	if err := c.BindJSON(&newVote); err != nil {
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate uuid: %v\n", err)
		os.Exit(1)
	}

	_, err = db.Exec(context.Background(), "INSERT INTO votes(id, target) VALUES($1, $2);", uuid, newVote.Target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Insert Row failed: %v\n", err)
		os.Exit(1)
	}

	c.IndentedJSON(http.StatusCreated, newVote)
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


	connection, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
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

	router := gin.Default()
	router.GET("/voteCount", getVoteCount)
	router.POST("/vote", postVote)

	router.Run("0.0.0.0:8080")
}
