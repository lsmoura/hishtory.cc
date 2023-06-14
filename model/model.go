package model

import "time"

type Device struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
	// The IP address that was used to register the device. Recorded so
	// that I can count how many people are using hishtory and roughly
	// from where. If you would like this deleted, please email me at
	// david@daviddworken.com and I can clear it from your device entries.
	RegistrationIP   string    `json:"registration_ip"`
	RegistrationDate time.Time `json:"registration_date"`
}
