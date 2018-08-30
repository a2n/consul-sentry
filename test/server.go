package main

import (
	"log"
	"net/http"

	"github.com/a2n/consul-sentry"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
	initZap()
	e := startDaemon()
	if e != nil {
		zap.L().Fatal("", zap.Error(e))
	}
}

func initZap() {
	l, e := zap.NewDevelopment()
	if e != nil {
		log.Fatalf("%+v", e)
	}
	zap.ReplaceGlobals(l)
}

func startDaemon() error {
	s := sentry.NewSentry()
	s.SetAddress(":8080")
	s.SetKeyFunc(key)
	s.SetErrorFunc(onError)
	ch := s.GetKeyCh()
	go func() {
		for {
			k, ok := <-ch
			if ok == false {
				return
			}
			zap.L().Debug(
				"goroutine",
				zap.Any("key", k),
			)
		}
	}()

	e := s.Start()
	if e != nil {
		return errors.Wrap(e, "")
	}

	return nil
}

func key(k *api.KVPair) {
	zap.L().Debug(
		"key",
		zap.Any("key", k),
	)
}

func onError(r *http.Request) {
	zap.L().Debug(
		"error",
		zap.Any("headers", r.Header),
	)
}
