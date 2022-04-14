package uievents

type EventType uint8

const (
	LoginSuccess = iota
	JoinSession
	SetToken
	RenewToken
	SessionStarted
)

type Event struct {
	Type    EventType
	Payload interface{}
	Error   error
}
