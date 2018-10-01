package main

import (
	"math/rand"
	"net/http"
	"os"

	"github.com/Scalingo/logrus-rollbar"
	"github.com/rollbar/rollbar-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v1"
)

func A() error {
	return errgo.New("error")
}

func B() error {
	return errgo.Mask(A(), errgo.Any)
}

func main() {
	rollbar.SetToken(os.Getenv("TOKEN"))
	rollbar.SetEnvironment("testing")

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
