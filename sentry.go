package sentry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Sentry Sentry
type Sentry struct {
	addr  string
	route string

	// callbacks
	keyFunc       KeyFunc
	keyPrefixFunc KeyPrefixFunc
	servicesFunc  ServicesFunc
	nodesFunc     NodesFunc
	serviceFunc   ServiceFunc
	checksFunc    ChecksFunc
	eventFunc     EventFunc
	errorFunc     ErrorFunc

	// channels
	rwMux        sync.RWMutex
	keyChs       map[chan *api.KVPair]struct{}
	keyPrefixChs map[chan api.KVPairs]struct{}
	servicesChs  map[chan map[string][]string]struct{}
	nodesChs     map[chan []*api.Node]struct{}
	serviceChs   map[chan []*api.ServiceEntry]struct{}
	checksChs    map[chan []*api.HealthCheck]struct{}
	eventChs     map[chan []*api.UserEvent]struct{}

	// indexs
	idx map[uint64]struct{}
}

// NewSentry Creating
func NewSentry() *Sentry {
	return &Sentry{
		keyChs:       map[chan *api.KVPair]struct{}{},
		keyPrefixChs: map[chan api.KVPairs]struct{}{},
		servicesChs:  map[chan map[string][]string]struct{}{},
		nodesChs:     map[chan []*api.Node]struct{}{},
		serviceChs:   map[chan []*api.ServiceEntry]struct{}{},
		checksChs:    map[chan []*api.HealthCheck]struct{}{},
		eventChs:     map[chan []*api.UserEvent]struct{}{},
		idx:          map[uint64]struct{}{},
	}
}

// SetAddress Setting web server address
func (s *Sentry) SetAddress(a string) {
	s.addr = a
}

// SetRoute Setting route.
func (s *Sentry) SetRoute(r string) {
	s.route = r
}

// Start Starting notifier
func (s *Sentry) Start() error {
	e := s.startHTTPDaemon()
	if e != nil {
		return errors.Wrap(e, "")
	}
	return nil
}

// startHTTPDaemon Starting HTTP daemon
func (s *Sentry) startHTTPDaemon() error {
	http.HandleFunc(s.route, s.handler)
	zap.L().Debug(
		"starting http daemon",
		zap.String("address", s.addr),
	)
	e := http.ListenAndServe(s.addr, nil)
	if e != nil {
		return errors.Wrap(e, "")
	}
	return nil
}

// handler HTTP handler
// Known bug, https://github.com/hashicorp/consul/issues/571
// Fires watch event two times.
func (s *Sentry) handler(w http.ResponseWriter, r *http.Request) {
	// idx, e := strconv.ParseUint(r.Header.Get("X-Consul-Index"), 10, 64)
	// if e != nil {
	// 	w.WriteHeader(http.StatusBadGateway)
	// 	zap.L().Error(
	// 		"no X-Consul-Index header",
	// 		zap.String("remote addr", r.RemoteAddr),
	// 	)
	// 	return
	// }
	// s.rwMux.RLock()
	// _, ok := s.idx[idx]
	// s.rwMux.RUnlock()
	// if ok == true {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }
	// s.rwMux.Lock()
	// s.idx[idx] = struct{}{}
	// s.rwMux.Unlock()

	t := strings.ToLower(r.Header.Get("type"))
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		zap.L().Error(
			"reading body",
			zap.Error(e),
		)
		return
	}
	if b == nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error("nil body")
		return
	}
	if len(b) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		zap.L().Error("empty body")
		return
	}

	switch t {
	case TypeKey.String():
		e = s.key(b)
	case TypeKeyPrefix.String():
		e = s.keyPrefix(b)
	case TypeServices.String():
		e = s.services(b)
	case TypeNodes.String():
		e = s.nodes(b)
	case TypeService.String():
		e = s.service(b)
	case TypeChecks.String():
		e = s.checks(b)
	case TypeEvent.String():
		e = s.event(b)
	default:
		if s.errorFunc != nil {
			s.errorFunc(r)
		}
	}
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error(
			"callback",
			zap.Error(e),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// KeyFunc Key function
type KeyFunc func(*api.KVPair)

// KeyPrefixFunc KeyPrefix function
type KeyPrefixFunc func(api.KVPairs)

// ServicesFunc Services function
type ServicesFunc func(map[string][]string)

// NodesFunc Nodes function
type NodesFunc func([]*api.Node)

// ServiceFunc Service function
type ServiceFunc func([]*api.ServiceEntry)

// ChecksFunc Checks function
type ChecksFunc func([]*api.HealthCheck)

// EventFunc Event function
type EventFunc func([]*api.UserEvent)

// ErrorFunc Error function
type ErrorFunc func(r *http.Request)

// SetKeyFunc Setting key function
func (s *Sentry) SetKeyFunc(fn KeyFunc) {
	s.keyFunc = fn
}

// GetKeyCh Returning key channel
func (s *Sentry) GetKeyCh() chan *api.KVPair {
	ch := make(chan *api.KVPair)
	s.rwMux.Lock()
	s.keyChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteKeyCh Deleting key channel
func (s *Sentry) DeleteKeyCh(ch chan *api.KVPair) {
	s.rwMux.Lock()
	delete(s.keyChs, ch)
	s.rwMux.Unlock()
}

// key key function
func (s *Sentry) key(b []byte) error {
	// unmarshal
	kv := new(api.KVPair)
	e := json.Unmarshal(b, &kv)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// notify
	if s.keyFunc != nil {
		s.keyFunc(kv)
	}

	s.rwMux.RLock()
	for k := range s.keyChs {
		k <- kv
	}
	s.rwMux.RUnlock()

	return nil
}

// SetKeyPrefixFunc Setting key prefix function
func (s *Sentry) SetKeyPrefixFunc(fn KeyPrefixFunc) {
	s.keyPrefixFunc = fn
}

// GetKeyPrefixCh Returning key prefix channel
func (s *Sentry) GetKeyPrefixCh() chan api.KVPairs {
	ch := make(chan api.KVPairs)
	s.rwMux.Lock()
	s.keyPrefixChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteKeyPrefixCh Deleting key prefix channel
func (s *Sentry) DeleteKeyPrefixCh(ch chan api.KVPairs) {
	s.rwMux.Lock()
	delete(s.keyPrefixChs, ch)
	s.rwMux.Unlock()
}

// keyPrefix key prefix function
func (s *Sentry) keyPrefix(b []byte) error {
	// unmarshal
	var kp api.KVPairs
	e := json.Unmarshal(b, &kp)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.keyPrefixFunc != nil {
		s.keyPrefixFunc(kp)
	}

	s.rwMux.RLock()
	for k := range s.keyPrefixChs {
		k <- kp
	}
	s.rwMux.RUnlock()

	return nil
}

// SetServicesFunc Setting services function
func (s *Sentry) SetServicesFunc(fn ServicesFunc) {
	s.servicesFunc = fn
}

// GetServicesCh Returning services channel
func (s *Sentry) GetServicesCh() chan map[string][]string {
	ch := make(chan map[string][]string)
	s.rwMux.Lock()
	s.servicesChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteServicesCh Deleting services channel
func (s *Sentry) DeleteServicesCh(ch chan map[string][]string) {
	s.rwMux.Lock()
	delete(s.servicesChs, ch)
	s.rwMux.Unlock()
}

// services services function
func (s *Sentry) services(b []byte) error {
	// unmarshal
	var ss map[string][]string
	e := json.Unmarshal(b, &ss)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.servicesFunc != nil {
		s.servicesFunc(ss)
	}

	s.rwMux.RLock()
	for k := range s.servicesChs {
		k <- ss
	}
	s.rwMux.RUnlock()

	return nil
}

// SetNodesFunc Setting nodes function
func (s *Sentry) SetNodesFunc(fn NodesFunc) {
	s.nodesFunc = fn
}

// GetNodesCh Returning nodes channel
func (s *Sentry) GetNodesCh() chan []*api.Node {
	ch := make(chan []*api.Node)
	s.rwMux.Lock()
	s.nodesChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteNodesCh Deleting nodes channel
func (s *Sentry) DeleteNodesCh(ch chan []*api.Node) {
	s.rwMux.Lock()
	delete(s.nodesChs, ch)
	s.rwMux.Unlock()
}

// node node function
func (s *Sentry) nodes(b []byte) error {
	// unmarshal
	var nodes []*api.Node
	e := json.Unmarshal(b, &nodes)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.nodesFunc != nil {
		s.nodesFunc(nodes)
	}

	s.rwMux.RLock()
	for k := range s.nodesChs {
		k <- nodes
	}
	s.rwMux.RUnlock()

	return nil
}

// SetServiceFunc Setting service function
func (s *Sentry) SetServiceFunc(fn ServiceFunc) {
	s.serviceFunc = fn
}

// GetServiceCh Returning service channel
func (s *Sentry) GetServiceCh() chan []*api.ServiceEntry {
	ch := make(chan []*api.ServiceEntry)
	s.rwMux.Lock()
	s.serviceChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteServiceCh Deleting service channel
func (s *Sentry) DeleteServiceCh(ch chan []*api.ServiceEntry) {
	s.rwMux.Lock()
	delete(s.serviceChs, ch)
	s.rwMux.Unlock()
}

// service service function
func (s *Sentry) service(b []byte) error {
	// unmarshal
	se := []*api.ServiceEntry{}
	e := json.Unmarshal(b, &se)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.serviceFunc != nil {
		s.serviceFunc(se)
	}

	s.rwMux.RLock()
	for k := range s.serviceChs {
		k <- se
	}
	s.rwMux.RUnlock()

	return nil
}

// SetChecksFunc Setting checks function
func (s *Sentry) SetChecksFunc(fn ChecksFunc) {
	s.checksFunc = fn
}

// GetChecksCh Returning nodes channel
func (s *Sentry) GetChecksCh() chan []*api.HealthCheck {
	ch := make(chan []*api.HealthCheck)
	s.rwMux.Lock()
	s.checksChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteChecksCh Deleting nodes channel
func (s *Sentry) DeleteChecksCh(ch chan []*api.HealthCheck) {
	s.rwMux.Lock()
	delete(s.checksChs, ch)
	s.rwMux.Unlock()
}

// checks checks function
func (s *Sentry) checks(b []byte) error {
	// unmarshal
	var hc []*api.HealthCheck
	e := json.Unmarshal(b, &hc)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.checksFunc != nil {
		s.checksFunc(hc)
	}

	s.rwMux.RLock()
	for k := range s.checksChs {
		k <- hc
	}
	s.rwMux.RUnlock()

	return nil
}

// SetEventFunc Setting event function
func (s *Sentry) SetEventFunc(fn EventFunc) {
	s.eventFunc = fn
}

// GetEventCh Returning event channel
func (s *Sentry) GetEventCh() chan []*api.UserEvent {
	ch := make(chan []*api.UserEvent)
	s.rwMux.Lock()
	s.eventChs[ch] = struct{}{}
	s.rwMux.Unlock()
	return ch
}

// DeleteEventCh Deleting event channel
func (s *Sentry) DeleteEventCh(ch chan []*api.UserEvent) {
	s.rwMux.Lock()
	delete(s.eventChs, ch)
	s.rwMux.Unlock()
}

// event event function
func (s *Sentry) event(b []byte) error {
	// unmarshal
	var ue []*api.UserEvent
	e := json.Unmarshal(b, &ue)
	if e != nil {
		return errors.Wrap(e, "")
	}

	// calling
	if s.eventFunc != nil {
		s.eventFunc(ue)
	}

	s.rwMux.RLock()
	for k := range s.eventChs {
		k <- ue
	}
	s.rwMux.RUnlock()

	return nil
}

// SetErrorFunc Setting error function
func (s *Sentry) SetErrorFunc(fn ErrorFunc) {
	s.errorFunc = fn
}
