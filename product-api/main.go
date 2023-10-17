package main

import (
	"context"
	"github.com/fleezesd/progressive-microservice/product-api/handler"
	"github.com/gorilla/mux"
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
	sm := mux.NewRouter().PathPrefix("/products").Subrouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("", ph.GetProducts)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("", ph.AddProduct)
	postRouter.Use(ph.MiddlewareValidateProduct)

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/{id:[0-9]+}", ph.UpdateProducts)
	putRouter.Use(ph.MiddlewareValidateProduct)

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
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
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
