package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/mxschmitt/golang-url-shortener/internal/handlers"
	"github.com/mxschmitt/golang-url-shortener/internal/stores"
	"github.com/mxschmitt/golang-url-shortener/internal/util"
	"github.com/pkg/errors"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

func main() {
	// Hack to get heroku vars in
	os.Setenv("GUS_LISTEN_ADDR", fmt.Sprintf("0.0.0.0:%v", os.Getenv("PORT")))

	rurl, err := url.Parse(os.Getenv("REDIS_URL"))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse redis url")
	}
	pw, _ := rurl.User.Password()
	os.Setenv("GUS_REDIS_HOST", rurl.Host)
	os.Setenv("GUS_REDIS_PASSWORD", pw)

	stop := make(chan os.Signal, 1)
	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}
	signal.Notify(stop, os.Interrupt)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   util.GetConfig().EnableColorLogs,
		DisableColors: !util.GetConfig().EnableColorLogs,
	})

	if util.GetConfig().EnableColorLogs == true {
		logrus.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	} else {
		logrus.SetOutput(os.Stdout)
	}
	close, err := initShortener()
	if err != nil {
		logrus.Fatalf("could not init shortener: %v", err)
	}
	<-stop
	logrus.Println("Shutting down...")
	close()
}

func initConfig() error {
	if err := util.ReadInConfig(); err != nil {
		return errors.Wrap(err, "could not load config")
	}
	return nil
}

func initShortener() (func(), error) {
	if util.GetConfig().EnableDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
	store, err := stores.New()
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler, err := handlers.New(*store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create handlers")
	}
	go func() {
		if err := handler.Listen(); err != nil {
			logrus.Fatalf("could not listen to http handlers: %v", err)
		}
	}()
	return func() {
		if err = handler.CloseStore(); err != nil {
			logrus.Printf("failed to stop the handlers: %v", err)
		}
	}, nil
}
