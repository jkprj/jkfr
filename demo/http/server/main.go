package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkhttp "github.com/jkprj/jkfr/gokit/transport/http"
	jklog "github.com/jkprj/jkfr/log"
)

type res struct {
	Action  string
	Content string
}

type Person struct {
	Name string
	Sex  string
	Age  int
}

func main() {
	jklog.InitLogger()

	runtime.GOMAXPROCS(runtime.NumCPU())

	mux := http.NewServeMux()
	mux.HandleFunc("/test", ServerHandle)

	// runDefautServer(mux)
	// runServerWithOption(mux)
	runServerWithDefaultConfigFile(mux)
	// runServerWithDefaultConfigFile2(mux)
	// runServerWithConfigFileOption(mux)
	// runMutiServer(mux)
}

func runDefautServer(handle http.Handler) {
	jkhttp.RunServerWithServerAddr("http", "192.168.213.184:8080", handle)
	// or
	// jkhttp.RunServer("http", handle, jkhttp.ServerAddr("192.168.213.184:8080"))
}

func runServerWithOption(handle http.Handler) {
	jkhttp.RunServer("http",
		handle,
		jkhttp.ServerAddr("127.0.0.1:8888"),
		jkhttp.ServerAddr("192.168.213.184:8080"), // 后面的会覆盖前面的设置
		jkhttp.ServerLimit(10),
		jkhttp.ServerRegOption(
			jkregistry.WithServerAddr("127.0.0.1:8888"), // 不会起作用，最终会设置为jkhttp.ServerAddr：192.168.213.184:8080
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
		),
	)
}

func runServerWithDefaultConfigFile(handle http.Handler) {
	// toml
	jkhttp.RunServer("http", handle)
	// or json
	// jkhttp.RunServer("http", handle)
}

func runServerWithConfigFileOption(handle http.Handler) {
	// toml
	// jkhttp.RunServer("http_tom", handle, jkhttp.ServerConfigFile("conf/httpT.toml"))
	// or json
	jkhttp.RunServer("http_json", handle, jkhttp.ServerConfigFile("conf/httpJ.json"))
}

func runMutiServer(handle http.Handler) {
	go jkhttp.RunServerWithServerAddr(
		"httpT",
		"127.0.0.1:8080",
		handle,
		jkhttp.ServerRegOption(
			jkregistry.WithTags("123", "jinkun"),
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
		),
	)

	jkhttp.RunServerWithServerAddr(
		"httpT",
		"192.168.213.184:8080",
		handle,
		jkhttp.ServerRegOption(
			jkregistry.WithTags("123", "jinkun"),
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
		),
	)
}
func ServerHandle(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("Action")
	if "SayHello" == action {
		re := res{Action: action + "Respone", Content: "hello baby, I'm " + r.Host}
		buf, _ := json.Marshal(re)
		w.Write(buf)
	} else if "HowAreYou" == action {
		re := res{Action: action + "Respone", Content: "fine, thank you, I'm " + r.Host}
		buf, _ := json.Marshal(re)
		w.Write(buf)
	} else if "WhatName" == action {
		buff, _ := ioutil.ReadAll(r.Body)
		jklog.Debugw("read body", "request_body", string(buff))
		ps := new(Person)
		json.Unmarshal(buff, ps)

		jklog.Debugw("person info", "person", *ps)

		re := res{Action: action + "Respone", Content: "Hello " + ps.Name + ", I'm " + r.Host}
		buf, _ := json.Marshal(re)
		w.Write(buf)

	} else {
		re := res{Action: action + "Respone", Content: "action not found"}
		buf, _ := json.Marshal(re)
		w.Write(buf)
	}
}
