package registry

import (
	"errors"
	"fmt"
	"sync"

	consul_api "github.com/hashicorp/consul/api"
)

var (
	ErrInvalidServiceRegistration = errors.New("service registration is invalid")
	ErrSyncing                    = errors.New("unable to sync with registry")
)

type Config struct {
	AdapterURI      string
	RefreshTTL      int
	RefreshInterval int
}

type ConsulAdapter struct {
	Offline   bool
	client    *consul_api.Client
	lastIndex uint64
	status    *AdapterStatus
	mtx       *sync.Mutex
}

type ServiceRegistration struct {
	Name             string
	Address          string
	Port             int
	Id               string
	Tags             []string
	TTL              string
	ConsulNodes      []string
	AdvertiseAddr    string
	SkipRegistration bool
}

type AdapterStatus struct {
	status int
}

func (as *AdapterStatus) setStatus(i int) {
	as.status = i
}

func (as *AdapterStatus) String() string {
	switch as.status {
	case 0:
		return "disconnected"
	case 1:
		return "connected"
	default:
		return "invalid"
	}
}

func (s *ServiceRegistration) String() string {
	return fmt.Sprintf("name: %s address: %s port: %v", s.Name, s.Address, s.Port)
}

func (s *ServiceRegistration) Valid() bool {
	if s.Name == "" {
		return false
	}
	if s.Address == "" {
		return false
	}
	if s.Port == 0 {
		return false
	}

	if s.Id == "" {
		return false
	}

	return true
}

type VailService struct {
	Name  string
	Nodes []*Node
}

type Node struct {
	*consul_api.CatalogService
}

type RegistryAdapter interface {
	Register(service ServiceRegistration) error
	DeRegister(service ServiceRegistration) error
	Sync(service ServiceRegistration) error

	Ping() error
	Status() int
	Disconnected() bool
	Type() string

	FindService(name, tag string) ([]*consul_api.CatalogService, error)
	FindServices() (map[string][]string, error)
	CheckService(name, tag string, passing bool) ([]*consul_api.ServiceEntry, error)
}
