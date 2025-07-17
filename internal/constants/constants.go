// Package constants defines shared constants used across the application
package constants

import "time"

const (
	// MinLenPassword defines the minimum allowed password length
	MinLenPassword = 6

	// MaxLenPassword defines the maximum allowed password length
	MaxLenPassword = 100

	// MinLenLogin defines the minimum allowed login length
	MinLenLogin = 3

	// MaxLenLogin defines the maximum allowed login length
	MaxLenLogin = 50

	// OneDayTimeout is used for cache/session expiration
	OneDayTimeout = 24 * time.Hour

	// ServerTimeout is read and write timeout of server config
	ServerTimeout = 15 * time.Second

	// DirPerm - Directory permission
	DirPerm = 0o750

	// FilePerm - File permission
	FilePerm = 0o600
)
