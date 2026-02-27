package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FPT-OJT/gateway/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestClientIp_XRealIP_TakesPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:9000"

	assert.Equal(t, "203.0.113.1", utils.ClientIp(req))
}

func TestClientIp_XForwardedFor_Single(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.5")

	assert.Equal(t, "198.51.100.5", utils.ClientIp(req))
}

func TestClientIp_XForwardedFor_MultipleIPs_ReturnsFirst(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 172.16.0.1, 192.168.1.1")

	assert.Equal(t, "10.0.0.1", utils.ClientIp(req))
}

func TestClientIp_RemoteAddr_WithPort_StripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:50123"

	assert.Equal(t, "192.0.2.10", utils.ClientIp(req))
}

func TestClientIp_RemoteAddr_NoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.99"

	assert.Equal(t, "192.0.2.99", utils.ClientIp(req))
}

func TestClientIp_NoHeaders_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Del("X-Real-IP")
	req.Header.Del("X-Forwarded-For")
	req.RemoteAddr = "1.2.3.4:8080"

	assert.Equal(t, "1.2.3.4", utils.ClientIp(req))
}
