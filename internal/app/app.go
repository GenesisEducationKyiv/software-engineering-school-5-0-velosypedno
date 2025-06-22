package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

const readTimeout = 15 * time.Second

type App struct {
	cfg    *config.Config
	db     *sql.DB
	cron   *cron.Cron
	apiSrv *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run() error {
	var err error

	// db
	a.db, err = sql.Open(a.cfg.DbDriver, a.cfg.DSN())
	if err != nil {
		return err
	}
	log.Println("DB connected")

	// cron
	a.cron = cron.New()
	err = setupCron(a.db, a.cron, a.cfg)
	if err != nil {
		return err
	}
	a.cron.Start()
	log.Println("Cron tasks are scheduled")

	// api
	router := gin.Default()
	setupRouter(a.db, router, a.cfg)
	a.apiSrv = &http.Server{
		Addr:        ":" + a.cfg.Port,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	go func() {
		if err := a.apiSrv.ListenAndServe(); err != nil {
			log.Printf("api server: %v", err)
		}
	}()
	log.Printf("APIServer started on port %s", a.cfg.Port)

	return nil
}

func (a *App) Shutdown(timeoutCtx context.Context) error {
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
