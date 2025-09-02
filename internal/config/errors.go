package config

import "errors"

// Configuration validation errors
var (
	ErrMissingTaskDir   = errors.New("task directory is required")
	ErrMissingServerURL = errors.New("server URL is required")
	ErrMissingAuth      = errors.New("username/password or token is required for authentication")
	ErrInvalidScanType  = errors.New("invalid scan type, must be one of: source, docker, binary")
	ErrInvalidThreadNum = errors.New("thread number must be between 1 and 60")
)
