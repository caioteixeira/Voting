package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func getTasks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, tasks)
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
	router := gin.Default()
	router.GET("/tasks", getTasks)
	router.POST("/tasks", postTask)

	router.Run("localhost:8080")
}
