package main

import (
	"math/rand"
	"net/http"
	"os"

	"github.com/Scalingo/logrus-rollbar"
	"github.com/Sirupsen/logrus"
	"github.com/stvp/rollbar"
	"gopkg.in/errgo.v1"
)

func A() error {
	return errgo.New("error")
}

func B() error {
	return errgo.Mask(A(), errgo.Any)
}

func main() {
	rollbar.Token = os.Getenv("TOKEN")
	rollbar.Environment = "testing"

	logger := logrus.New()
	logger.Hooks.Add(logrus_rollbar.Hook{})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := B()

		logger.WithFields(
			logrus.Fields{"req": r, "error": err, "extra-data": rand.Int()},
		).Error("Something is really wrong")
	})

	http.ListenAndServe(":31313", nil)
}
