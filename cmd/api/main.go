package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"mdb/internal/data"
	"mdb/internal/jsonlog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	env  string
	port string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct{
		enable bool
		rps float64
		burst int 
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config
	flag.StringVar(&cfg.port, "port", "4000", "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")
	// flag.StringVar(&cfg.db.dsn, "dsn",os.Getenv("MDB_DSN"), "PSQL DSN")
	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://mdb:pa55word@localhost/mdb", "PSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Maximun Open DB Connection")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Maximun Idle DB Connection")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Maximun Idle time for a DB Connection")
	flag.IntVar(&cfg.limiter.burst, "limiter-brust", 4, "Rate Limiter max brust")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate Limiter max rate per minute")
	flag.BoolVar(&cfg.limiter.enable,"limiter-enabled",true,"Enable rate limiter")
	flag.Parse()
	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	db, errDB := openDB(cfg)
	if errDB != nil {
		log.Fatal(errDB)
	}
	defer db.Close()
	logf, errf := os.OpenFile("mdb.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if errf != nil {
		log.Fatalln(errf.Error())
	}
	jLogger := jsonlog.New([]io.Writer{os.Stdout, logf}, jsonlog.LevelInfo)
	app := &application{
		config: cfg,
		logger: jLogger,
		models: data.NewModel(db),
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(jLogger, "", 0),
	}
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": cfg.port,
		"env":  cfg.env,
	})
	// logger.Printf("starting %s server on %v port.", cfg.env, cfg.port)
	// err := r.Run(fmt.Sprintf(":%d", cfg.port))
	err := srv.ListenAndServe()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(c)
	if err != nil {
		return nil, err
	}
	log.Println("Ping to DB was success")
	return db, nil
}
