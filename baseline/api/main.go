package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

var tasks = []task{
	{ID: "1", Description: "Potato", Completed: false},
	{ID: "2", Description: "Onion", Completed: true},
}

type entry struct {
	ID    string `json:"id"`
	Votes int    `json:"votes"`
}

var db *pgxpool.Pool

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

func postTask(c *gin.Context) {
	var newTask task

	if err := c.BindJSON(&newTask); err != nil {
		return
	}

	tasks = append(tasks, newTask)
	c.IndentedJSON(http.StatusCreated, newTask)
}

func main() {
	var err error
	db, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	router := gin.Default()
	router.GET("/voteCount", getVoteCount)
	router.POST("/tasks", postTask)

	router.Run("localhost:8080")
}
