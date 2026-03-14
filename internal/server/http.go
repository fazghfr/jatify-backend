package server

import (
	"github.com/gin-gonic/gin"
)

func NewEngine() *gin.Engine {
	r := gin.Default()
	return r
}

func Run(r *gin.Engine, port string) error {
	return r.Run(":" + port)
}
