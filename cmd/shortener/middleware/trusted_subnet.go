package middleware

import (
	"net"
	"net/http"
	"sync"
)

// TrustedSubnetCache кеширует распарсенные подсети
type TrustedSubnetCache struct {
	subnet string
	net    *net.IPNet
	mutex  sync.RWMutex
}

var subnetCache = &TrustedSubnetCache{}

// TrustedSubnetMiddleware проверяет, что IP адрес клиента находится в доверенной подсети
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если доверенная подсеть не настроена, запрещаем доступ
			if trustedSubnet == "" {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Получаем IP адрес из заголовка X-Real-IP
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				// Если заголовок X-Real-IP отсутствует, используем RemoteAddr
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					clientIP = r.RemoteAddr
				} else {
					clientIP = host
				}
			}

			// Получаем кешированную подсеть или парсим новую
			trustedNet, err := getCachedSubnet(trustedSubnet)
			if err != nil {
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			// Парсим IP адрес клиента
			clientIPAddr := net.ParseIP(clientIP)
			if clientIPAddr == nil {
				http.Error(w, "Invalid client IP", http.StatusBadRequest)
				return
			}

			// Проверяем, входит ли IP адрес в доверенную подсеть
			if !trustedNet.Contains(clientIPAddr) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getCachedSubnet возвращает кешированную подсеть или парсит новую
func getCachedSubnet(subnet string) (*net.IPNet, error) {
	subnetCache.mutex.RLock()
	if subnetCache.subnet == subnet && subnetCache.net != nil {
		net := subnetCache.net
		subnetCache.mutex.RUnlock()
		return net, nil
	}
	subnetCache.mutex.RUnlock()

	// Парсим подсеть
	_, trustedNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	// Кешируем результат
	subnetCache.mutex.Lock()
	subnetCache.subnet = subnet
	subnetCache.net = trustedNet
	subnetCache.mutex.Unlock()

	return trustedNet, nil
}
