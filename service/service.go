package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mailgun/manners"
	"gitlab.vailsys.com/vail-cloud-services/platform"
	"gitlab.vailsys.com/vail-cloud-services/platform/heimdal"
	"gitlab.vailsys.com/vail-cloud-services/platform/middleware"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
)

const (
	// Interval between consul heartbeats
	defaultRegisterInterval = 2 * time.Second
)

var (
	ErrInvalidHandler        = errors.New("invalid service handler")
	ErrInvalidHandlerMethods = errors.New("invalid service handler methods")
)

type Service struct {
	Registration    registry.ServiceRegistration
	Router          *gin.Engine
	ServiceHandlers []ServiceHandler
	RegistryAdapter registry.RegistryAdapter
	ServiceClients  []heimdal.HttpServiceClient
	Logger          *logrus.Logger
	pulse           *registry.Pulse
	mtx             *sync.Mutex
	srv             *manners.GracefulServer
}

func NewService(registration registry.ServiceRegistration) *Service {

	router := gin.New()

	service := &Service{
		Registration:    registration,
		Router:          router,
		ServiceHandlers: make([]ServiceHandler, 0),
		ServiceClients:  make([]heimdal.HttpServiceClient, 0),
		mtx:             &sync.Mutex{},
		Logger:          platform.Logger,
	}

	//router.Use(requestId(service.Name()))
	router.Use(serviceLogger())

	service.initRegistry()

	addr := fmt.Sprintf("%v:%v", service.Registration.Address, service.Registration.Port)
	srv := manners.NewWithServer(&http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	service.srv = srv

	return service
}

func (service *Service) initRegistry() error {
	nodes := service.Registration.ConsulNodes
	var node string

	if len(nodes) == 0 {
		node = "consul://127.0.0.1:8500"
	} else {
		node = nodes[0]
	}

	config := registry.Config{
		AdapterURI:      node,
		RefreshTTL:      5,
		RefreshInterval: 20,
	}

	adapter, err := registry.NewBackend(config)
	if err != nil {
		service.Logger.Infof("registry backend error: %s", err)
	}

	service.RegistryAdapter = adapter
	return nil
}

func (service *Service) register() error {
	if service.RegistryAdapter.Disconnected() {
		return fmt.Errorf("registry adapter %s is not connected", service.RegistryAdapter.Type())
	}

	if service.Synced() {
		return fmt.Errorf("service %s is already registered", service.Name())
	}

	err := service.RegistryAdapter.Register(service.Registration)

	if err != nil {
		return err
	}

	service.Logger.Infof("service %s registered with consul", service.Registration.Name)

	pulser, err := registry.NewPulser(defaultRegisterInterval, service.Registration, service.RegistryAdapter)

	if err != nil {
		return err
	}

	service.pulse = pulser

	return nil
}

func (service *Service) AddHandler(sh ServiceHandler) error {
	if sh.Handler == nil {
		return ErrInvalidHandler
	}

	if len(sh.Methods) == 0 {
		return ErrInvalidHandlerMethods
	}

	service.addHandlers(&service.Router.RouterGroup, sh)

	return nil
}

func (service *Service) AddGroupedHandler(group string, sh ServiceHandler) error {
	if sh.Handler == nil {
		return ErrInvalidHandler
	}

	if len(sh.Methods) == 0 {
		return ErrInvalidHandlerMethods
	}

	gHandler := service.Router.Group(group)

	service.addHandlers(gHandler, sh)

	return nil
}

func (service *Service) SetNotFoundHandler(h gin.HandlerFunc) error {
	if h == nil {
		return ErrInvalidHandler
	}
	service.Router.NoRoute(h)
	return nil
}

func (service *Service) AddMiddleware(m middleware.Middleware) error {
	handler := m.GinFunc()
	if handler == nil {
		return fmt.Errorf("invalid middleware handler")
	}
	service.Router.Use(handler)
	return nil
}

func (service *Service) AddGroupMiddleware(group string, m middleware.Middleware) error {
	handler := m.GinFunc()
	if handler == nil {
		return fmt.Errorf("invalid middleware handler")
	}
	gHandler := service.Router.Group(group)
	gHandler.Use(handler)
	return nil
}

func (service *Service) AddServiceClient(c heimdal.HttpServiceClient) {
	_, err := service.GetServiceClient(c.ServiceName)
	if err != nil {
		service.ServiceClients = append(service.ServiceClients, c)
	}
}

func (service *Service) GetServiceClient(name string) (*heimdal.HttpServiceClient, error) {
	service.Logger.Debugf("finding service for name: %s", name)
	for _, c := range service.ServiceClients {
		if c.ServiceName == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("service client with name: %s does not exist", name)
}

func (service *Service) Run() error {
	service.Logger.Infof("running service %s", service.Registration.Name)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	if !service.Registration.SkipRegistration {
		err := service.register()

		if err != nil {
			service.Logger.Infof("service %s is not registered", service.Name())
			return err
		}
		go func() {
			service.pulse.Start()
		}()
	}

	//sigexit code
	go func(e chan os.Signal) {
		s := <-e
		service.Logger.Infof("Received signal: %v, shutting down", s)
		service.Stop()
	}(exitChan)

	return service.srv.ListenAndServe()
}

func (service *Service) Stop() {
	if service.pulse != nil {
		service.pulse.Stop()
	}
	service.stopServiceClients()
	service.srv.Close()
	//service.srv.BlockingClose()
}

func (service *Service) Name() string {
	service.mtx.Lock()
	defer service.mtx.Unlock()
	return service.Registration.Name
}

func (service *Service) Synced() bool {
	service.mtx.Lock()
	defer service.mtx.Unlock()

	if service.pulse == nil {
		return false
	}

	return service.pulse.Active()
}

func (service *Service) addHandlers(router *gin.RouterGroup, sh ServiceHandler) {
	for _, path := range sh.Paths {
		for _, method := range sh.Methods {
			service.Logger.Debugf("adding handler to service router %s: %s", method, path)
			router.Handle(method, path, sh.Handler)
		}
	}

	service.ServiceHandlers = append(service.ServiceHandlers, sh)
}

func (service *Service) stopServiceClients() {
	for _, c := range service.ServiceClients {
		service.Logger.Debugf("stoping loadbalancer for %s", c.ServiceName)
		c.Loadbalancer.Stop()
	}
}

func serviceLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId, _ := c.Get("requestid")

		path := c.Request.URL.Path
		platform.Logger.Infof("requestid=%s method=%s path=%s agent=%s host=%s request=%s", requestId.(string), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Host, c.Request.RequestURI)

		start := time.Now()

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		method := c.Request.Method
		statusCode := c.Writer.Status()

		platform.Logger.Infof("requestid=%s status=%d latency=%v method=%s path=%s", requestId.(string), statusCode, latency, method, path)
	}
}

//func requestId(name string) gin.HandlerFunc {
//m := middleware.NewRequestID(name)
//return m.GinFunc()
//}
