package main

import (
	"context"
	"github.com/fleezesd/progressive-microservice/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	l := log.New(os.Stdout, "product-api ", log.Lshortfile|log.LstdFlags)
	hh := handler.NewHello(l)
	sm := http.NewServeMux()
	sm.Handle("/", hh)

	s := http.Server{
		Addr:        ":9090",
		Handler:     sm,
		IdleTimeout: 1 * time.Second,
		ReadTimeout: 1 * time.Second,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			l.Fatal(err)
		}
	}()
	// 识别signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)

	// 优雅关闭
	tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.Shutdown(tc)
}
