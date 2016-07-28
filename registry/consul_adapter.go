package registry

import (
	"fmt"
	"net/url"
	"sync"

	"gitlab.vailsys.com/vail-cloud-services/platform"
	"gitlab.vailsys.com/vail-cloud-services/platform/heimdal"

	consul_api "github.com/hashicorp/consul/api"
)

const (
	StatusDisconnected = 0
	StatusConnected    = 1
	CONSUL_TYPE        = "consul"
)

func NewConsulAdapter(uri *url.URL) RegistryAdapter {
	config := consul_api.DefaultConfig()
	config.HttpClient = heimdal.DefaultHttpClient()

	if uri.Host != "" {
		config.Address = uri.Host
	}

	client, err := consul_api.NewClient(config)
	if err != nil {
		fmt.Printf("error creating consul adapter %s", err)
	}
	status := &AdapterStatus{status: StatusDisconnected}

	adapter := &ConsulAdapter{client: client, status: status, mtx: &sync.Mutex{}}
	adapter.Ping()

	return adapter
}

func (c *ConsulAdapter) Leader() (*string, error) {
	status := c.client.Status()

	leader, err := status.Leader()
	if err != nil {
		return nil, err
	}
	return &leader, nil
}

func (c *ConsulAdapter) Register(sr ServiceRegistration) error {
	if sr.TTL == "" {
		sr.TTL = "5s"
	}

	service := &consul_api.AgentServiceRegistration{
		Address: sr.AdvertiseAddr,
		Port:    sr.Port,
		ID:      sr.Id,
		Name:    sr.Name,
		Tags:    sr.Tags,
		Check:   c.createTTLCheck(sr),
	}

	agent := c.client.Agent()

	err := agent.ServiceRegister(service)
	if err != nil {
		return err
	}

	platform.Logger.Debugf("registering service %s", sr.String())

	return nil
}

func (c *ConsulAdapter) DeRegister(sr ServiceRegistration) error {
	agent := c.client.Agent()

	err := agent.ServiceDeregister(sr.Id)

	if err != nil {
		return err
	}

	platform.Logger.Debugf("deregistering service %s", sr.String())
	return nil
}

func (c *ConsulAdapter) Sync(sr ServiceRegistration) error {
	key := c.createCheckKey(sr.Id)
	agent := c.client.Agent()

	err := agent.PassTTL(key, "pass")
	if err != nil {
		return err
	}
	return nil
}

// Ping will try to connect to consul by attempting to retrieve the current leader.
func (c *ConsulAdapter) Ping() error {
	status := c.client.Status()
	leader, err := status.Leader()

	if err != nil {
		c.setStatus(StatusDisconnected)
		return err
	}

	platform.Logger.Debugf("consul current leader %s", leader)

	peers, _ := status.Peers()
	platform.Logger.Debugf("consul current peers: %s", peers)

	c.setStatus(StatusConnected)

	return nil
}

func (c *ConsulAdapter) Status() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.status.status
}

func (c *ConsulAdapter) FindServices() (map[string][]string, error) {
	catalog := c.client.Catalog()
	qo := &consul_api.QueryOptions{AllowStale: false, RequireConsistent: true, WaitIndex: c.lastIndex}

	services, meta, err := catalog.Services(qo)

	if err != nil {
		return nil, err
	}

	c.lastIndex = meta.LastIndex

	platform.Logger.Debugf("consul meta %s", meta)

	return services, nil
}

func (c *ConsulAdapter) FindService(name, tag string) ([]*consul_api.CatalogService, error) {
	catalog := c.client.Catalog()
	qo := &consul_api.QueryOptions{AllowStale: false, RequireConsistent: true}
	//, WaitIndex: c.lastIndex}

	cnodes, meta, err := catalog.Service(name, tag, qo)
	if err != nil {
		return nil, err
	}

	if len(cnodes) == 0 {
		return nil, fmt.Errorf("service %s not found", name)
	}

	c.lastIndex = meta.LastIndex

	platform.Logger.Debugf("consul meta %s", meta)

	return cnodes, nil
}

func (c *ConsulAdapter) CheckService(name, tag string, passing bool) ([]*consul_api.ServiceEntry, error) {
	_, err := c.FindService(name, "")

	if err != nil {
		return nil, err
	}

	qo := &consul_api.QueryOptions{AllowStale: false, RequireConsistent: true}
	//, WaitIndex: c.lastIndex}

	health := c.client.Health()

	entries, meta, err := health.Service(name, tag, passing, qo)
	if err != nil {
		return nil, err
	}

	platform.Logger.Debugf("consul meta %s", meta)
	c.lastIndex = meta.LastIndex

	return entries, nil
}

func (c *ConsulAdapter) Disconnected() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.status.status == StatusDisconnected {
		return true
	}
	return false
}

func (c *ConsulAdapter) Type() string {
	return CONSUL_TYPE
}

func (c *ConsulAdapter) createTTLCheck(sr ServiceRegistration) *consul_api.AgentServiceCheck {
	return &consul_api.AgentServiceCheck{TTL: sr.TTL}
}

func (c *ConsulAdapter) createCheckKey(id string) string {
	return "service:" + id
}

func (c *ConsulAdapter) setStatus(status int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if status != c.status.status {
		platform.Logger.Debugf("setting status for adapter %s", c.status.String())
		c.status.setStatus(status)
	}
}
