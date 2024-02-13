package main

import (
	"fmt"
	"olra-v1/internal/server"
	// "github.com/gin-contrib/cors"
)

func main() {

	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
