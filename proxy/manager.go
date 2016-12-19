package proxy

import (
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/janus/router"
)

const (
	methodAll = "ALL"
)

// Register represents a register proxy
type Register struct {
	Router    router.Router
	proxies   []Proxy
	Transport http.RoundTripper
}

// RegisterMany registers many proxies at once
func (p *Register) RegisterMany(proxies []Proxy, handlers ...router.Constructor) {
	for _, proxy := range proxies {
		p.Register(proxy, handlers...)
	}
}

// Register register a new proxy
func (p *Register) Register(proxy Proxy, handlers ...router.Constructor) {
	if false == p.Exists(proxy) {
		handler := p.ToHandler(proxy)
		matcher := router.NewListenPathMatcher()
		if matcher.Match(proxy.ListenPath) {
			p.doRegister(matcher.Extract(proxy.ListenPath), handler, proxy.Methods, handlers)
		}

		p.doRegister(proxy.ListenPath, handler, proxy.Methods, handlers)
		p.proxies = append(p.proxies, proxy)
	}
}

func (p *Register) doRegister(
	listenPath string,
	handler http.HandlerFunc,
	methods []string,
	handlers []router.Constructor,
) {
	log.WithFields(log.Fields{
		"listen_path": listenPath,
	}).Info("Registering a proxy")

	for _, method := range methods {
		if strings.ToUpper(method) == methodAll {
			p.Router.Any(listenPath, handler, handlers...)
		} else {
			p.Router.Handle(strings.ToUpper(method), listenPath, handler, handlers...)
		}
	}
}

// Exists checks if a proxy is already registered in the manager
func (p *Register) Exists(proxy Proxy) bool {
	for _, route := range p.proxies {
		if route.ListenPath == proxy.ListenPath {
			return true
		}
	}

	return false
}

// ToHandler turns a proxy configuration into a handler
func (p *Register) ToHandler(proxy Proxy) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		reverseProxy := NewSingleHostReverseProxy(proxy, p.Transport)
		reverseProxy.ServeHTTP(rw, r)
	}
}