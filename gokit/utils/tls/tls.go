package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func CreateTLSListen(server_pem []byte, server_key []byte, client_pem []byte, linkAddr string) (net.Listener, error) {

	cert, err := tls.X509KeyPair(server_pem, server_key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(client_pem)
	if !ok {
		errorInfo := "failed to parse root certificate"
		log.Printf(errorInfo)
		return nil, errors.New(errorInfo)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCertPool,
	}

	ln, err := tls.Listen("tcp", linkAddr, config)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ln, nil
}

func CreateTLSConn(client_pem []byte, client_key []byte, linkAddr string, timeOut time.Duration) (conn *tls.Conn, err error) {

	cert, err := tls.X509KeyPair(client_pem, client_key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(client_pem)
	if !ok {
		errorInfo := fmt.Sprintf("failed to parse root certificate")
		log.Printf(errorInfo)
		return nil, errors.New(errorInfo)
	}

	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	if -1 == timeOut {
		conn, err = tls.Dial("tcp", linkAddr, conf)
	} else {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeOut}, "tcp", linkAddr, conf)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}
