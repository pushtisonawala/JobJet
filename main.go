package main

import (
	"fmt"
	"jobqueue/controllers"
	"jobqueue/db"
	"jobqueue/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	logger.Log.Info("application starting")
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	fmt.Println("server started")
	db.ConnectMongo()
	r.POST("/jobs", controllers.CreateJobs)
	r.Run(":8000")
}
