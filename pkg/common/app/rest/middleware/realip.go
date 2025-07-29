package middleware

import (
	"net"
	"net/http"
	"strings"
)

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

func RealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if raddr := realAddr(r); raddr != "" {
			r.RemoteAddr = raddr
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

/*
		@property
	    def remote_ip(self):
	        """We don't use Tornado ``HTTPServerRequest.remote_ip`` directly, cause
	        this value depend on ``HTTPServer.xheaders`` setting. Gunicorn recent
	        change (http://docs.gunicorn.org/en/latest/news.html#tornado-worker)
	        make us cannot set this setting. So we first get remote IP from header.
	        """
	        # get remote IP from XFF hdr, fallback to request.remote_ip
	        remote_ip = self.request.headers.get("X-Forwarded-For",
	                                             self.request.remote_ip)
	        remote_ip = remote_ip.split(',')[0].strip()

	        # prefer to get remote IP from X-Real-IP hdr, fallback to XFF IP
	        remote_ip = self.request.headers.get("X-Real-Ip", remote_ip)
	        remote_ip = remote_ip.split(',')[0].strip()

	        if is_valid_ip(remote_ip):
	            return remote_ip
	        return self.request.remote_ip
*/
func realAddr(r *http.Request) string {
	// The HTTP server in this package sets RemoteAddr to an "IP:port" address
	// before invoking a handler.
	remoteIP, remotePort, _ := net.SplitHostPort(r.RemoteAddr)

	if forwardFor := r.Header.Get(xForwardedFor); forwardFor != "" {
		ips := strings.Split(forwardFor, ",")
		if len(ips) > 0 {
			s := ips[len(ips)-1]
			remoteIP = strings.Trim(s, " \n\t")
		}
	}

	if realIP := r.Header.Get(xRealIP); realIP != "" {
		remoteIP = realIP
	}
	if !isValidIp(remoteIP) {
		return r.RemoteAddr
	}
	return net.JoinHostPort(remoteIP, remotePort)
}

func isValidIp(ip string) bool {
	return net.ParseIP(ip) != nil
}
