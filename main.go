package main

import (
	"service"

	flag "github.com/spf13/pflag"
)

const (
	// PORT 默认监听端口
	PORT string = "12380"
)

func main() {
	var port string
	pPort := flag.StringP("port", "p", PORT, "PORT for httpd listening")
	flag.Parse()
	if len(*pPort) != 0 {
		port = *pPort
	} else {
		port = PORT
	}

	server := service.NewServer()
	server.Run(":" + port)
}
