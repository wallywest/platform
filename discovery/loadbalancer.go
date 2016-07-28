package discovery

import (
	"errors"
	"net/url"
)

type LoadBalancer interface {
	Count() int
	Get() (*url.URL, error)
	Stop()
}

var ErrNoEndpointsAvailable = errors.New("no endpoints available")
