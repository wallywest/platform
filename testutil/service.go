package testutil

import (
	"gitlab.vailsys.com/vail-cloud-services/platform"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
	"gitlab.vailsys.com/vail-cloud-services/platform/service"
)

func NewMockRegistration() registry.ServiceRegistration {
	consulNodes := []string{"consul://127.0.0.1:8500"}
	name := "testservice"
	registration := registry.ServiceRegistration{
		Name:        "testservice",
		Address:     "127.0.0.1",
		Port:        20000,
		TTL:         "15s",
		Id:          platform.GenerateUUID(name),
		ConsulNodes: consulNodes,
	}
	return registration
}

func NewMockService() *service.Service {
	reg := NewMockRegistration()
	service := service.NewService(reg)
	return service
}
