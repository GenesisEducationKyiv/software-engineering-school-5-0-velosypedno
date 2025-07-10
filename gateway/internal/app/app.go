package app

import (
	"context"
	"fmt"
	"log"
	"time"
)

const shutdownTimeout = 20 * time.Second

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context) error {
Loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down signal received")
			break Loop
		default:
			fmt.Println("Hello")
			time.Sleep(time.Second)
		}
	}

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := a.shutdown(timeoutCtx)
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	return shutdownErr
}
