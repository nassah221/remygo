package flags

import (
	"fmt"
	"log"
	"strings"

	app "github.com/remygo/application"

	"github.com/pion/webrtc/v3"
)

func validateURL(cfg *app.Args) error {
	if strings.HasPrefix(cfg.URL, "turn:") {
		if cfg.TurnCreds == "" {
			return fmt.Errorf("provide TURN credentials with the '--creds' flag")
		}
		if !strings.ContainsAny(cfg.TurnCreds, ":") {
			return fmt.Errorf("provide TURN credentials in the format 'username:password'")
		}
	} else if strings.HasPrefix(cfg.URL, "stun:") {
		if cfg.TurnCreds != "" {
			return fmt.Errorf("do not provide TURN credentials with the '--creds' flag when using STUN")
		}
	}

	return nil
}

// func validateMode(cfg *app.Args) error {
// 	if cfg.Mode == "" {
// 		return fmt.Errorf("provide a mode with the '-mode' flag i.e. 'host' or 'remote'")
// 	}
// 	switch cfg.Mode {
// 	case "host":
// 		log.Println("[INFO] Running as host")
// 	case "remote":
// 		if cfg.Token == "" {
// 			return fmt.Errorf("provide a token with the '-token' flag when running as remote")
// 		}
// 	}

// 	return nil
// }

func validateCodec(cfg *app.Args) error {
	// if cfg.Mode == "remote" {
	// 	log.Println("[INFO] Host peer will choose the codec")
	// 	return nil
	// }
	switch cfg.Codec {
	case webrtc.MimeTypeH264:
		log.Println("[INFO] Using H264 codec")
	case webrtc.MimeTypeVP8:
		log.Println("[INFO] Using VP8 codec")
	case webrtc.MimeTypeVP9:
		log.Println("[INFO] Using VP9 codec")
	default:
		return fmt.Errorf("[ERR] Unsupported codec: %s", cfg.Codec)
	}
	return nil
}

// Validates the flags provided by the user needed to start the application.
// Does not validate user credentials
func Validate(cfg *app.Args) error {
	if err := validateURL(cfg); err != nil {
		return err
	}

	// if err := validateMode(cfg); err != nil {
	// 	return err
	// }

	// Check if codec is one of the supported ones
	return validateCodec(cfg)
}
