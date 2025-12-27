package main

import (
	"fmt"
	"jobqueue/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	fmt.Println("server started")
	r.POST("/jobs", controllers.CreateJobs)
	r.Run(":8000")
}
