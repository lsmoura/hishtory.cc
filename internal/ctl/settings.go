package ctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/client"
	"github.com/rs/xid"
	"os"
)

type Settings struct {
	Hostname   string             `json:"hostname"`
	DeviceID   string             `json:"device_id"`
	UserSecret *client.UserSecret `json:"user_id"`

	fn string `json:"-"`
}

func ReadSettings(fn string) (Settings, error) {
	ret := Settings{
		fn: fn,
	}

	r, err := os.Open(fn)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ret, nil
		}
		return ret, fmt.Errorf("os.ReadSettings: %w", err)
	}

	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return ret, fmt.Errorf("json.NewDecoder.Decode: %w", err)
	}

	return ret, nil
}

func (s *Settings) Save(fn string) error {
	if fn == "" {
		fn = s.fn
	}

	if s.UserSecret == nil {
		secret, err := client.NewUserSecret()
		if err != nil {
			return fmt.Errorf("client.NewUserSecret: %w", err)
		}
		s.UserSecret = secret
	}

	w, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}

	if err := json.NewEncoder(w).Encode(s); err != nil {
		return fmt.Errorf("json.NewEncoder.Encode: %w", err)
	}

	return nil
}

func (s *Settings) GetDeviceID() (string, error) {
	if s.DeviceID == "" {
		s.DeviceID = xid.New().String()

		if err := s.Save(""); err != nil {
			return "", fmt.Errorf("cannot save settings file: %w", err)
		}
	}

	return s.DeviceID, nil
}
