package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rinha-backend-go/api"
	"rinha-backend-go/persistence/postgres"

	"golang.org/x/sync/errgroup"
)

func main() {

	dsn := os.Getenv("DSN")

	if dsn == "" {
		log.Fatal("DSN environment variable not set")
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT environment variable not set")
	}

	redisAddress := os.Getenv("REDIS_ADDRESS")

	if redisAddress == "" {
		log.Fatal("REDIS_ADDRESS environment variable not set")
	}

	store, err := postgres.NewPostgresStore(dsn)

	if err != nil {
		log.Fatal(err)
	}

	server := api.New(store, "8080", redisAddress)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(context.Background())

	log.Println("Starting server")
	g.Go(server.Start)

	select {
	case <-ctx.Done():
		err = server.Stop()
		if err != nil {
			log.Println("Error:", err)
			os.Exit(1)
		}
	case sig := <-interrupt:
		log.Println("Received signal:", sig)
		err = server.Stop()
		if err != nil {
			log.Println("Error:", err)
			os.Exit(1)
		}
	}

	log.Println("Shutting down")

	if err := g.Wait(); err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}

}
