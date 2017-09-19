package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/Scalingo/logrus-rollbar"
	"github.com/sirupsen/logrus"
	"github.com/stvp/rollbar"
)

func main() {
	rollbar.Token = os.Getenv("TOKEN")
	rollbar.Environment = "testing"

	logger := logrus.New()
	logger.Hooks.Add(logrus_rollbar.Hook{})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := fmt.Errorf("something wrong happened in the database")

		logger.WithFields(
			logrus.Fields{"req": r, "error": err, "extra-data": rand.Int()},
		).Error("Something is really wrong")
	})

	http.ListenAndServe(":31313", nil)
}
