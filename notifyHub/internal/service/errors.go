package service

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDeliveryFailed = errors.New("delivery failed")
)
