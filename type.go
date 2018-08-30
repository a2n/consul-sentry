package sentry

// WatchType Watch type
type WatchType int

const (
	// TypeKey Key type
	TypeKey WatchType = iota

	// TypeKeyPrefix KeyPrefix type
	TypeKeyPrefix

	// TypeServices Services type
	TypeServices

	// TypeNodes Nodes type
	TypeNodes

	// TypeService Service type
	TypeService

	// TypeChecks Checks type
	TypeChecks

	// TypeEvent Event type
	TypeEvent
)

func (t WatchType) String() string {
	var s string
	switch t {
	case TypeKey:
		s = "key"
	case TypeKeyPrefix:
		s = "keyprefix"
	case TypeServices:
		s = "services"
	case TypeNodes:
		s = "nodes"
	case TypeService:
		s = "service"
	case TypeChecks:
		s = "checks"
	case TypeEvent:
		s = "event"
	}
	return s
}
