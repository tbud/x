package embryo

import (
	"net"
	"net/http/httputil"
)

type Embryo struct {
	serverHost string
	port       int
	proxy      *httputil.ReverseProxy
}

func NewEmbryo() (embryo *Embryo) {
	embryo = &Embryo{}

	return
}

func getFreePort() (port int) {
	conn, err := net.Listen("tcp", ":0")
	if err != nil {

	}
}
