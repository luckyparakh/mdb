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
	"mdb/internal/mailer"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var version string
var buildTime string

type config struct {
	env  string
	port string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		enable bool
		rps    float64
		burst  int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.StringVar(&cfg.port, "port", "4000", "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")
	// flag.StringVar(&cfg.db.dsn, "dsn",os.Getenv("MDB_DSN"), "PSQL DSN")
	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://mdb:pa55word@localhost/mdb", "PSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Maximun Open DB Connection")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Maximun Idle DB Connection")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Maximun Idle time for a DB Connection")
	flag.IntVar(&cfg.limiter.burst, "limiter-brust", 4, "Rate Limiter max brust")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate Limiter max rate per minute")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "127.0.0.1", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "MDB <noreply@mdb.rp.net>", "SMTP sender")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}
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
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	customMetric(db)
	err := app.server()
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
