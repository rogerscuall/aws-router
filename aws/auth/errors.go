package auth

import "errors"

var (
	ErrNoDefaultAuthentication = errors.New("no default authentication was found")
	ErrStsIdentityNotFound     = errors.New("sts identity not found")
)
