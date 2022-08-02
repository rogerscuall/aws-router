package auth

import "errors"

var (
	ErrNoDefaultAuthentication = errors.New("no default authentication was found")
	ErrSTSIdentityNotFound     = errors.New("sts identity not found")
	ErrNoEC2ProfileRole		= errors.New("no ec2 profile role was found")
)
