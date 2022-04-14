package handler

import (
	"context"
	"log"

	"github.com/remygo/swagger"
)

type LoggingEvent uint8

const (
	CreateSession LoggingEvent = iota
	EndSession
	JoinSession
	LeaveSession
)

type APICall struct {
	Type         LoggingEvent
	DeviceID     string
	UserID       string
	SessionToken string
}

// Logging service is a goroutine that listens to the API channel and logs
// the events to the REST API
func (m *Manager) LoggingService(apiChan chan APICall) {
	c := swagger.NewAPIClient(swagger.NewConfiguration())

	for call := range apiChan {
		switch call.Type {
		case CreateSession:
			_, _, err := c.SessionApi.Create(context.TODO(), call.DeviceID, swagger.AddSession{Identifier: call.SessionToken})
			if err != nil {
				log.Printf("[API] Error in call %s: %v", call.Type.String(), err)
			}
		case EndSession:
			_, _, err := c.SessionApi.EndSessionById(context.TODO(), call.SessionToken)
			if err != nil {
				log.Printf("[API] Error in call %s: %v", call.Type.String(), err)
			}
		case JoinSession:
			_, _, err := c.SessionDeviceApi.Create(context.TODO(), call.SessionToken, call.DeviceID)
			if err != nil {
				log.Printf("[API] Error in call %s: %v", call.Type.String(), err)
			}
		case LeaveSession:
			_, _, err := c.SessionDeviceApi.EndDeviceSession(context.TODO(), call.SessionToken, call.DeviceID)
			if err != nil {
				log.Printf("[API] Error in call %s: %v", call.Type.String(), err)
			}
		}
	}
}

func (l LoggingEvent) String() string {
	switch l {
	case CreateSession:
		return "CreateSession"
	case EndSession:
		return "EndSession"
	case JoinSession:
		return "JoinSession"
	case LeaveSession:
		return "LeaveSession"
	default:
		return "Unsupported"
	}
}
