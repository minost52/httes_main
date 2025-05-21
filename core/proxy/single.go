package proxy

import (
	"net/url"
)

const ProxyTypeSingle = "single"

func init() {
	AvailableProxyServices[ProxyTypeSingle] = &singleProxyStrategy{}
}

type singleProxyStrategy struct {
	proxyAddr *url.URL
}

func (sp *singleProxyStrategy) Init(p Proxy) error {
	sp.proxyAddr = p.Addr
	return nil
}

func (sp *singleProxyStrategy) GetAll() []*url.URL {
	return []*url.URL{sp.proxyAddr}
}

func (sp *singleProxyStrategy) GetProxy() *url.URL {
	return sp.proxyAddr
}

func (sp *singleProxyStrategy) ReportProxy(addr *url.URL, reason string) *url.URL {
	return sp.proxyAddr
}

func (sp *singleProxyStrategy) GetProxyCountry(addr *url.URL) string {
	return "unknown"
}

func (sp *singleProxyStrategy) Done() error {
	return nil
}
