package main

import (
	"context"
	"github.com/fleezesd/progressive-microservice/product-api/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	l := log.New(os.Stdout, "product-api ", log.Lshortfile|log.LstdFlags)
	// create handlers
	ph := handler.NewProducts(l)

	// create a new serve mux and register handlers
	sm := http.NewServeMux()
	sm.Handle("/products", ph)

	// create a new server
	s := http.Server{
		Addr:        ":9090",
		Handler:     sm,
		IdleTimeout: 1 * time.Second,
		ReadTimeout: 1 * time.Second,
	}

	// server start
	go func() {
		if err := s.ListenAndServe(); err != nil {
			l.Fatal(err)
		}
	}()

	// trap sigterm or interrupt signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)

	// gracefully shutdown server
	tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.Shutdown(tc)
}
