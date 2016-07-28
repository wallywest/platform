package heimdal

import (
	"io"
	"net/http"
	"time"

	"github.com/facebookgo/httpcontrol"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery"
)

var DEFAULT_RETRIES = uint(0)
var DEFAULT_TIMEOUT = 5 * time.Second

type HttpRequestBuilderOptions struct {
	ID      string
	Path    string
	Method  string
	Payload io.ReadCloser
}

type requestFunc func(*http.Request, func())
type responseFunc func(*http.Response, func())

type PayloadFunc func(*http.Request) (io.ReadCloser, error)

type HttpRequestBuilder struct {
	ID           string
	Path         string
	Method       string
	Payload      io.ReadCloser
	Headers      map[string]string
	Values       map[string]string
	RequestFunc  []requestFunc
	ResponseFunc []responseFunc
}

type HttpServiceClient struct {
	ServiceName  string
	Loadbalancer discovery.LoadBalancer
	Client       *http.Client
}

var DEFAULT_CONTENT_TYPE = "application/json"

func NewHttpServiceClient(name string, lb discovery.LoadBalancer) HttpServiceClient {
	client := DefaultHttpClient()
	return HttpServiceClient{
		ServiceName:  name,
		Loadbalancer: lb,
		Client:       client,
	}
}

func (h *HttpServiceClient) Execute(builder *HttpRequestBuilder) (*http.Response, error) {

	url, err := h.Loadbalancer.Get()
	if err != nil {
		return nil, err
	}

	url.Path = builder.Path

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(builder.Method, url.String(), builder.Payload)

	if err != nil {
		return nil, err
	}

	for k, v := range builder.Headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", DEFAULT_CONTENT_TYPE)
	val := req.URL.Query()

	for k, v := range builder.Values {
		val.Set(k, v)
	}
	req.URL.RawQuery = val.Encode()

	//before
	for _, fn := range builder.RequestFunc {
		fn(req, func() {})
	}

	//request
	resp, err := h.Client.Do(req)

	if err != nil {
		return nil, err
	}

	//after
	for _, fn := range builder.ResponseFunc {
		fn(resp, func() {})
	}

	return resp, nil
}

func NewHttpRequestBuilder(options HttpRequestBuilderOptions) *HttpRequestBuilder {
	return &HttpRequestBuilder{
		ID:           options.ID,
		Method:       options.Method,
		Path:         options.Path,
		Payload:      options.Payload,
		Headers:      map[string]string{},
		RequestFunc:  make([]requestFunc, 0),
		ResponseFunc: make([]responseFunc, 0),
	}
}

func (b *HttpRequestBuilder) AddReqFunc(r requestFunc) {
	b.RequestFunc = append(b.RequestFunc, r)
}

func (b *HttpRequestBuilder) AddRespFunc(w responseFunc) {
	b.ResponseFunc = append(b.ResponseFunc, w)
}

func (b *HttpRequestBuilder) SetQueryString(qs map[string]string) {
	if qs != nil {
		b.Values = qs
	}
}

func (b *HttpRequestBuilder) AddHeader(k, v string) {
	ok := b.Headers[k]
	if ok == "" {
		b.Headers[k] = v
	}
}

type TransportOptions struct {
	*httpcontrol.Transport
}

func DefaultHttpClient() *http.Client {
	c := &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout:    DEFAULT_TIMEOUT,
			MaxTries:          DEFAULT_RETRIES,
			DisableKeepAlives: true,
		},
	}
	return c
}

func NewHttpClient(options *TransportOptions) *http.Client {
	if options == nil {
		return DefaultHttpClient()
	}

	tr := options.Transport
	c := &http.Client{
		Transport: tr,
	}
	return c
}
