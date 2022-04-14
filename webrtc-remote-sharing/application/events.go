package application

type EventType uint8

const (
	Renew EventType = iota
	InSession
	SessionEnded
)

type SessionEvent struct {
	Type EventType
}
