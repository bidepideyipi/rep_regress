package middleware

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/platform-games/gateway/pkg/response"
)

type IPWhitelistMiddleware struct {
	allowedNets []*net.IPNet
}

func NewIPWhitelistMiddleware(allowedIPs []string) (*IPWhitelistMiddleware, error) {
	m := &IPWhitelistMiddleware{
		allowedNets: make([]*net.IPNet, 0, len(allowedIPs)),
	}

	for _, cidr := range allowedIPs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			ip := net.ParseIP(cidr)
			if ip == nil {
				return nil, err
			}

			if ip.To4() != nil {
				ipNet = &net.IPNet{
					IP:   ip.To4(),
					Mask: net.CMask(net.IPv4len),
				}
			} else {
				ipNet = &net.IPNet{
					IP:   ip.To16(),
					Mask: net.CMask(net.IPv6len),
				}
			}
		}

		m.allowedNets = append(m.allowedNets, ipNet)
	}

	return m, nil
}

func (m *IPWhitelistMiddleware) Middleware() gin.HandlerFunc {
	if len(m.allowedNets) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			response.Forbidden(c, "Invalid client IP")
			c.Abort()
			return
		}

		allowed := false
		for _, ipNet := range m.allowedNets {
			if ipNet.Contains(clientIP) {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "IP not in whitelist")
			c.Abort()
			return
		}

		c.Next()
	}
}
