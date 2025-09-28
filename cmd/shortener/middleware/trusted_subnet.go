package middleware

import (
	"net"
	"net/http"
)

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

			// Парсим доверенную подсеть
			_, trustedNet, err := net.ParseCIDR(trustedSubnet)
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
