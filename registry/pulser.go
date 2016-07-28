package registry

import (
	"fmt"
	"sync"
	"time"

	"gitlab.vailsys.com/vail-cloud-services/platform"
)

type Pulse struct {
	active       bool
	interval     time.Duration
	adapter      RegistryAdapter
	registration ServiceRegistration
	ticker       *time.Ticker
	mtx          *sync.RWMutex
	quit         chan int
}

func NewPulser(interval time.Duration, registration ServiceRegistration, adapter RegistryAdapter) (*Pulse, error) {
	if !registration.Valid() {
		return nil, ErrInvalidServiceRegistration
	}

	//fix this to check ttl
	dur, _ := time.ParseDuration(registration.TTL)

	if interval > dur {
		return nil, fmt.Errorf("must use pulse interval: %s greater then registration TTL %s", interval.String(), registration.TTL)
	}

	return &Pulse{interval: interval, registration: registration, adapter: adapter, mtx: &sync.RWMutex{}}, nil
}

func (p *Pulse) Start() {
	platform.Logger.Debugf("Starting heartbeat for app: %v", p.registration)
	p.setStatus(true)
	p.ticker = time.NewTicker(p.interval)
	p.quit = make(chan int)
	go p.Beat()
}

func (p *Pulse) Stop() {
	platform.Logger.Debugf("Stopping heartbeat for app: %v", p.registration)
	close(p.quit)
	p.ticker.Stop()
	p.adapter.DeRegister(p.registration)
	p.setStatus(false)
}

func (p *Pulse) Beat() {
	for {
		select {
		case <-p.ticker.C:
			//date race
			if p.adapter.Status() == StatusDisconnected {
				platform.Logger.Debugf("registry adapter connection unavailable")
				p.adapter.Ping()
				return
			}
			err := p.adapter.Sync(p.registration)
			if err != nil {
				platform.Logger.Infof("pulser beat has stopped %s", err)
			}
		case <-p.quit:
			platform.Logger.Infof("quiting the pulser beat")
			return
		}
	}
}

func (p *Pulse) setStatus(val bool) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	p.active = val
}

func (p *Pulse) Active() bool {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.active
}
