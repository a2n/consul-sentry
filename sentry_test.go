package sentry

import (
	"log"
	"testing"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

func initZap() {
	l, e := zap.NewDevelopment()
	if e != nil {
		log.Fatalf("%+v", e)
	}
	zap.ReplaceGlobals(l)
}

func TestKey(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `{"Key":"foo/bar/baz","CreateIndex":1793,"ModifyIndex":1793,"LockIndex":0,"Flags":0,"Value":"aGV5","Session":""}`
	s.SetKeyFunc(func(k *api.KVPair) {
		if k.Key != "foo/bar/baz" {
			t.Errorf("mismatch key, %s", k.Key)
		}
	})
	go func() {
		for k := range s.GetKeyCh() {
			if k.Key != "foo/bar/baz" {
				t.Errorf("mismatch key, %s", k.Key)
			}
		}
	}()

	s.key([]byte(raw))
}

func TestKeyPrefix(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `[{"Key":"foo/bar","CreateIndex":1796,"ModifyIndex":1796,"LockIndex":0,"Flags":0,"Value":"TU9BUg==","Session":""},{"Key":"foo/baz","CreateIndex":1795,"ModifyIndex":1795,"LockIndex":0,"Flags":0,"Value":"YXNkZg==","Session":""},{"Key":"foo/test","CreateIndex":1793,"ModifyIndex":1793,"LockIndex":0,"Flags":0,"Value":"aGV5","Session":""}]`
	s.SetKeyPrefixFunc(func(k api.KVPairs) {
		if k[0].Key != "foo/bar" {
			t.Errorf("mismatch key, %s", k[0].Key)
		}
	})
	go func() {
		for k := range s.GetKeyPrefixCh() {
			if k[0].Key != "foo/bar" {
				t.Errorf("mismatch key, %s", k[0].Key)
			}
		}
	}()

	s.keyPrefix([]byte(raw))
}

func TestServices(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `{"consul":[],"redis":[],"web":[]}`
	s.SetServicesFunc(func(m map[string][]string) {
		_, ok := m["consul"]
		if ok == false {
			t.Errorf("consul not exists")
		}
	})
	go func() {
		for m := range s.GetServicesCh() {
			_, ok := m["consul"]
			if ok == false {
				t.Errorf("consul not exists")
			}
		}
	}()

	s.services([]byte(raw))
}

func TestNodes(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `[{"Node":"nyc1-consul-1","Address":"192.241.159.115"},{"Node":"nyc1-consul-2","Address":"192.241.158.205"},{"Node":"nyc1-consul-3","Address":"198.199.77.133"},{"Node":"nyc1-worker-1","Address":"162.243.162.228"},{"Node":"nyc1-worker-2","Address":"162.243.162.226"},{"Node":"nyc1-worker-3","Address":"162.243.162.229"}]`
	s.SetNodesFunc(func(ns []*api.Node) {
		if ns[0].Node != "nyc1-consul-1" {
			t.Errorf("Key is not \"nyc1-consul-1\"")
		}
	})
	go func() {
		for ns := range s.GetNodesCh() {
			if ns[0].Node != "nyc1-consul-1" {
				t.Errorf("Key is not \"nyc1-consul-1\"")
			}
		}
	}()

	s.nodes([]byte(raw))
}

func TestService(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `[{"Node":{"Node":"foobar","Address":"10.1.10.12"},"Service":{"ID":"redis","Service":"redis","Tags":null,"Port":8000},"Checks":[{"Node":"foobar","CheckID":"service:redis","Name":"Service 'redis' check","Status":"passing","Notes":"","Output":"","ServiceID":"redis","ServiceName":"redis"},{"Node":"foobar","CheckID":"serfHealth","Name":"Serf Health Status","Status":"passing","Notes":"","Output":"","ServiceID":"","ServiceName":""}]}]`
	s.SetServiceFunc(func(ss []*api.ServiceEntry) {
		if ss[0].Node.Node != "foobar" {
			t.Errorf("Node is not foobar")
		}
	})
	go func() {
		for ss := range s.GetServiceCh() {
			if ss[0].Node.Node != "foobar" {
				t.Errorf("Node is not foobar")
			}
		}
	}()

	s.service([]byte(raw))
}

func TestChecks(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `[{"Node":"foobar","CheckID":"service:redis","Name":"Service 'redis' check","Status":"passing","Notes":"","Output":"","ServiceID":"redis","ServiceName":"redis"}]`
	s.SetChecksFunc(func(cs []*api.HealthCheck) {
		if cs[0].Node != "foobar" {
			t.Errorf("Node is not foobar")
		}
	})
	go func() {
		for cs := range s.GetChecksCh() {
			if cs[0].Node != "foobar" {
				t.Errorf("Node is not foobar")
			}
		}
	}()

	s.checks([]byte(raw))
}

func TestEvent(t *testing.T) {
	initZap()
	s := NewSentry()
	raw := `[{"ID":"f07f3fcc-4b7d-3a7c-6d1e-cf414039fcee","Name":"web-deploy","Payload":"MTYwOTAzMA==","NodeFilter":"","ServiceFilter":"","TagFilter":"","Version":1,"LTime":18}]`
	s.SetEventFunc(func(es []*api.UserEvent) {
		if es[0].ID != "f07f3fcc-4b7d-3a7c-6d1e-cf414039fcee" {
			t.Errorf("invalid ID")
		}
	})
	go func() {
		for es := range s.GetEventCh() {
			if es[0].ID != "f07f3fcc-4b7d-3a7c-6d1e-cf414039fcee" {
				t.Errorf("invalid ID")
			}
		}
	}()

	s.event([]byte(raw))
}
