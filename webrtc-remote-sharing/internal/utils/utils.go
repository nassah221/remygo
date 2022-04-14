package utils

import (
	"encoding/json"
	"log"

	"github.com/go-vgo/robotgo"
)

func NormalizedPos(x, y, remoteW, remoteH float32, userW, userH int) (nX, nY int) {
	nX = int((x / remoteW) * float32(userW))
	nY = int((y / remoteH) * float32(userH))
	return
}

func MarshalEvent(event interface{}) (eventJSON []byte) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Println("Unable to marshal remote event: ", err)
	}
	return
}

func GetDisplaySize() (width, height int) {
	// TODO: Handle multiple display dimensions properly

	// Get the display scale factor
	scale := robotgo.ScaleF()
	width, height = robotgo.GetScreenSize()

	// Normalize the dimensions as GetScreenSize returns the scaled resolution
	width, height = int(float64(width)*scale), int(float64(height)*scale)

	return
}
