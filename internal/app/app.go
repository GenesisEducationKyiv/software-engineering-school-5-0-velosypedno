package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

const readTimeout = 15 * time.Second
const shutdownTimeout = 20 * time.Second
const logFilepath = "log.log"
const logPerm os.FileMode = 0644

type App struct {
	cfg         *config.Config
	db          *sql.DB
	cron        *cron.Cron
	apiSrv      *http.Server
	reposLogger *log.Logger
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	f, err := os.OpenFile(logFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logPerm)
	if err != nil {
		return err
	}
	a.reposLogger = log.New(f, "", log.LstdFlags)

	// db
	a.db, err = sql.Open(a.cfg.DB.Driver, a.cfg.DB.DSN())
	if err != nil {
		return err
	}
	log.Println("DB connected")

	// cron
	err = a.setupCron()
	if err != nil {
		return err
	}
	a.cron.Start()
	log.Println("Cron tasks are scheduled")

	// api
	router := a.setupRouter()
	a.apiSrv = &http.Server{
		Addr:        ":" + a.cfg.Srv.Port,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	go func() {
		if err := a.apiSrv.ListenAndServe(); err != nil {
			log.Printf("api server: %v", err)
		}
	}()
	log.Printf("APIServer started on port %s", a.cfg.Srv.Port)

	// wait on shutdown signal
	<-ctx.Done()

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = a.shutdown(timeoutCtx)
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// api
	if a.apiSrv != nil {
		if err := a.apiSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown api server: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("APIServer Shutdown successfully")
		}
	}

	// cron
	if a.cron != nil {
		log.Println("Stopping cron scheduler")
		cronCtx := a.cron.Stop()
		select {
		case <-cronCtx.Done():
			log.Println("Cron scheduler stopped")
		case <-timeoutCtx.Done():
			wrapped := fmt.Errorf("shutdown cron scheduler: %w", timeoutCtx.Err())
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		}
	}

	// db
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown db: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("DB closed")
		}
	}
	return shutdownErr
}
