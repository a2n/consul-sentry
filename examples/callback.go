package main

import (
	"log"

	sentry "github.com/a2n/consul-sentry"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

func main() {
	initZap()
	s := sentry.NewSentry()
	s.SetAddress(":8080")
	s.SetKeyFunc(key)
	e := s.Start()
	if e != nil {
		log.Fatalf("%+v", e)
	}
}

func key(k *api.KVPair) {
	log.Printf("%s: %s", k.Key, k.Value)
}

func initZap() {
	l, e := zap.NewDevelopment()
	if e != nil {
		panic(e)
	}
	zap.ReplaceGlobals(l)
}
