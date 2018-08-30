package main

import (
	"log"

	sentry "github.com/a2n/consul-sentry"
	"go.uber.org/zap"
)

func main() {
	initZap()
	s := sentry.NewSentry()
	s.SetAddress(":8080")

	ch := s.GetKeyCh()
	go func() {
		for pair := range ch {
			log.Printf("%s: %s", pair.Key, pair.Value)
		}
	}()

	e := s.Start()
	if e != nil {
		log.Fatalf("%+v", e)
	}
}

func initZap() {
	l, e := zap.NewDevelopment()
	if e != nil {
		panic(e)
	}
	zap.ReplaceGlobals(l)
}
