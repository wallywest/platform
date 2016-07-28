package registry

import (
	"fmt"
	"net/url"
)

func NewBackend(config Config) (RegistryAdapter, error) {
	uri, err := url.Parse(config.AdapterURI)
	if err != nil {
		return nil, fmt.Errorf("Invalid adapter URI %v", err)
	}

	switch uri.Scheme {
	case "consul":
		adapter := NewConsulAdapter(uri)
		return adapter, nil
	default:
		return nil, fmt.Errorf("Invalid adapter scheme %v", uri.Scheme)
	}
}
