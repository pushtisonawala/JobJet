package main

import (
	"fmt"
	"jobqueue/controllers"
	"jobqueue/db"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// Set trusted proxies for Gin (removes warning)
	r.SetTrustedProxies([]string{"127.0.0.1"})

	fmt.Println("server started")
	db.ConnectMongo()
	r.POST("/jobs", controllers.CreateJobs)
	r.Run(":8000")
}
