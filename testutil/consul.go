package testutil

import (
	"io/ioutil"
	"testing"

	consul_test "github.com/hashicorp/consul/testutil"
)

type ConsulCluster struct {
	Leader *consul_test.TestServer
	Agents []*consul_test.TestServer
}

func NewConsulCluster(t *testing.T) ConsulCluster {
	cluster := ConsulCluster{}

	srv := consul_test.NewTestServerConfig(t, func(c *consul_test.TestServerConfig) {
		c.Stderr = ioutil.Discard
		c.Stdout = ioutil.Discard
		c.LogLevel = "err"
	})

	srv2 := consul_test.NewTestServerConfig(t, func(c *consul_test.TestServerConfig) {
		c.Stderr = ioutil.Discard
		c.Stdout = ioutil.Discard
		c.LogLevel = "err"
		c.Bootstrap = false
	})

	srv.JoinLAN(srv2.LANAddr)

	cluster.Leader = srv
	cluster.Agents = append(cluster.Agents, srv2)

	return cluster
}

func (c ConsulCluster) Stop() {
	c.Leader.Stop()
	for _, a := range c.Agents {
		a.Stop()
	}
}
