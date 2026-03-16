package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// New создаёт обработчик, который перенаправляет запросы на targetURL
func New(targetURL string) http.Handler {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Опционально: можно модифицировать запрос перед отправкой (например, прокинуть X-User-ID)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	
	return proxy
}
