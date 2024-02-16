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
	// server.GET("https://wema-alatdev-apimgt.azure-api.net/callback-url/callbackURL", controllers.CallBack)

	controllers.UserRoutes(basepath)
	return server
}
