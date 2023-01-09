package rpc

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	jktrans "jkfr/gokit/transport"
	jklog "jkfr/log"
)

type Server struct {
	rpc.Server
	codec string
}

func NewServer(codec string) *Server {

	s := new(Server)
	s.codec = codec

	return s
}

// 重写 rpc.Server 的 ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		jklog.Errorw("rpc hijacking fail", "RemoteAddr", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")

	if jktrans.CODEC_JSON == s.codec {
		s.ServeCodec(jsonrpc.NewServerCodec(conn))
	} else {
		s.ServeConn(conn)
	}
}
