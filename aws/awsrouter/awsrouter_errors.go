package awsrouter

import "errors"

var (
	ErrTgwRouteTableNotFound = errors.New("awsrouter: transit gateway route table not found")
	ErrTgwRouteTableRouteNotFound = errors.New("awsrouter: transit gateway route table route not found")
	ErrTgwAttachmetInPath = errors.New("awsrouter: attachmet is already in the path")
)
