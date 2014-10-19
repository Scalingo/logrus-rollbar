# rollbar hook for [Logrus](https://github.com/Sirupsen/logrus)

## Setup

```sh
go get github.com/Appsdeck/logrus-rollbar
```

## Example

```go
package main

import (
	"fmt"
	"net/http"
	"math/rand"
	
	"github.com/Sirupsen/logrus"
	rollbar "github.com/AlekSi/rollbar-go"
	logrus_rollbar "github.com/Appsdeck/logrus-rollbar"
)


func main() {
	rollbar.ApiKey = "123456ABCD"
	rollbar.Environment = "testing"

	logger := logrus.New()
	logger.Hooks.Add(logrus_rollbar.Hook{})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := fmt.Errorf("something wrong happened in the database")

		logger.WithFields(
			logrus.Fields{"req": r, "error": err, "extra-data", rand.Int()},
		).Error("Something is really wrong")
	})

	http.ListenAndServe(":31313", nil)
}
```
