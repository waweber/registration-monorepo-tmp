package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DBURL string
}

type Server struct {
	pool *pgxpool.Pool
}

func RunServer(config Config, port int) {
	pool, err := pgxpool.New(context.Background(), config.DBURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	server := Server{
		pool: pool,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /events/{eventId}/registrations", server.HandleCreateRegistration)
	mux.HandleFunc("GET /events/{eventId}/registrations/{registrationId}", server.HandleGetRegistration)
	mux.HandleFunc("PUT /events/{eventId}/registrations/{registrationId}", server.HandleUpdateRegistration)

	log.Printf("listening on %v", port)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
