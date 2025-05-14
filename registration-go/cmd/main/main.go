package main

import (
	"flag"
	"registration/internal/server"
)

func main() {
	var dburl string
	var port int

	flag.StringVar(&dburl, "db", "postgres://127.0.0.1/reg", "database URL")
	flag.IntVar(&port, "port", 8000, "port to listen on")

	flag.Parse()

	cfg := server.Config{
		DBURL: dburl,
	}

	server.RunServer(cfg, port)
}
