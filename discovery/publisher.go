package discovery

import "net/url"

type Publisher interface {
	Subscribe(chan<- []*url.URL)
	Unsubscribe(chan<- []*url.URL)
	Stop()
}
