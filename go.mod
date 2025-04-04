module github.com/Scalingo/logrus-rollbar

go 1.22
toolchain go1.24.1

require (
	github.com/Scalingo/errgo-rollbar v0.2.1
	github.com/pkg/errors v0.9.1
	github.com/rollbar/rollbar-go v1.4.6
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/errgo.v1 v1.0.1
)

require golang.org/x/sys v0.31.0 // indirect
