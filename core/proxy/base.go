package proxy

import (
	"fmt"
	"net/url"
	"reflect"
)

var AvailableProxyServices = make(map[string]ProxyService)

// Структура прокси-сервера используется для инициализации реализаций прокси-сервиса.
type Proxy struct {
	// Стратегия использования прокси-сервера.
	Strategy string
	// Установить это поле, если стратегия прокси-сервера является единой
	Addr *url.URL
	// Динамическое поле для других прокси-стратегий.
	Others map[string]interface{}
}

// ProxyService - это интерфейс, который абстрагирует различные реализации прокси.
// Поле Strategy в типах.Прокси определяет, какую реализацию использовать.
type ProxyService interface {
	Init(Proxy) error
	GetAll() []*url.URL
	GetProxy() *url.URL
	ReportProxy(addr *url.URL, reason string) *url.URL
	GetProxyCountry(*url.URL) string
	Done() error
}

// NewProxyService - это заводской метод работы прокси-сервиса.
func NewProxyService(s string) (service ProxyService, err error) {
	if val, ok := AvailableProxyServices[s]; ok {
		// Создайте новый объект из типа сервиса
		service = reflect.New(reflect.TypeOf(val).Elem()).Interface().(ProxyService)
	} else {
		err = fmt.Errorf("unsupported proxy strategy: %s", s)
	}

	return
}
