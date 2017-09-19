package main

import (
	"math/rand"
	"net/http"
	"os"

	"github.com/Scalingo/logrus-rollbar"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/stvp/rollbar"
)

func A() error {
	return errors.New("error")
}

func B() error {
	return errors.Wrap(A(), "A failed")
}

func main() {
	rollbar.Token = os.Getenv("TOKEN")
	rollbar.Environment = "staging"

	logger := logrus.New()
	logger.Hooks.Add(logrus_rollbar.New(0))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := B()

		logger.WithFields(
			logrus.Fields{"req": r, "error": err, "extra-data": rand.Int()},
		).Error("Something is really wrong")
	})

	http.ListenAndServe(":31313", nil)
}
