package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) server() error {
	shutDownErr := make(chan error)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.PrintInfo("Shutdown initiated", map[string]string{"signal": s.String()})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.logger.PrintInfo("Completing Background tasks", map[string]string{"signal": s.String()})
		app.wg.Wait()
		shutDownErr <- srv.Shutdown(ctx)
	}()
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": app.config.port,
		"env":  app.config.env,
	})
	// err := r.Run(fmt.Sprintf(":%d", cfg.port))
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	sdErr := <-shutDownErr
	if sdErr != nil {
		return sdErr
	}
	app.logger.PrintInfo("Shutdown completed", map[string]string{"addr": app.config.port, "env": app.config.env})
	return nil
}
