package main

import (
	"fmt"
	"jobqueue/db"
	"jobqueue/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	fmt.Println("server started")
	db.ConnectMongo()
	r.POST("/jobs", controllers.CreateJobs)
	r.Run(":8000")
}
