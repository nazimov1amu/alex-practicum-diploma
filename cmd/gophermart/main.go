package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nazimov1amu/alex-practicum-diploma/internal/app"
)

func main() {
	application, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		if err := application.Stop(); err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}
}
