package main

import (
	"fmt"
	"olra-v1/internal/server"

	// "github.com/gin-contrib/cors"

	helpers "olra-v1/utils"
)

func main() {

	secretKey, errr := helpers.GenerateRandomString(32)
	if errr != nil {
		fmt.Println("Error generating secret key:", errr)
		return
	}
	fmt.Println("Random secret key:", secretKey)

	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
