package server

import (
	"net/http"
	"olra-v1/controllers" // Import your controllers package

	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func (s *Server) RegisterRoutes() http.Handler {
	server = gin.Default()
	basepath := server.Group("/api/v1")

	controllers.UserRoutes(basepath)
	return server
}
