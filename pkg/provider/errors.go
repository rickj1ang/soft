package provider

import "errors"

var (
	ErrEmptyResponse     = errors.New("empty response from provider")
	ErrEmptyBase64       = errors.New("empty base64 string")
	ErrModelNotSupported = errors.New("model not supported")
	ErrEmptyMessages     = errors.New("empty messages")
)
