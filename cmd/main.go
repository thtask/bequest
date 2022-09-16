package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotunj/bequest/config"
	"github.com/dotunj/bequest/internal/pkg/app"
	"github.com/dotunj/bequest/internal/pkg/server"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	interrupt = make(chan os.Signal, 1)
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	logrus.SetFormatter(&prefixed.TextFormatter{
		DisableColors:   false,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceFormatting: true,
	})

	var mongoDsn, redisDsn, port string

	flag.StringVar(&mongoDsn, "mongo-dsn", "", "MongoDB DSN")
	flag.StringVar(&port, "port", "", "Server Port")

	flag.Parse()

	//Set up Config
	cfg, err := config.NewConfig(mongoDsn, redisDsn, port)
	if err != nil {
		logrus.Fatal(err)
	}

	//Create a new application
	app, err := app.NewApplication(cfg.Database.Dsn)
	if err != nil {
		logrus.Fatal(err)
	}

	//close DB connection
	defer app.DB.DB.Client().Disconnect(context.Background())

	httpServer := server.New(app.Routes(), cfg.Server.Port)

	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logrus.Infof("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		logrus.Errorf("app - Run - httpServer.Notify: %v", err)
	}

	err = httpServer.Shutdown()
	if err != nil {
		logrus.Fatal(fmt.Errorf("app - Run - httpServer.Shutdown: %v", err))
	}
}
