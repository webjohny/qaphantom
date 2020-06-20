package main

import (
	"fmt"
	"github.com/cssivision/reverseproxy"
	"net/http"
	"net/url"
)

var proxies []Proxy

type Proxy struct {
	Host string
	Port string
	Login string
	Password string

	LocalIp string
	Log []string
}

func (p *Proxy) NetProxy() string {
	//if p.Host != ""{
	//	return ""
	//}
	proxyAddr, _ := url.ParseRequestURI("http://1SCq278p:nkAXhVlq@45.89.19.37:12410")
	pxy := &reverseproxy.ReverseProxy{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyAddr),},
		Director: func(req *http.Request) {
			req.URL = proxyAddr
		},
	}
	go func() {
		err := http.ListenAndServe("localhost:9090", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "CONNECT" {
				pxy.ProxyHTTPS(w, r)
			} else {
				pxy.ProxyHTTP(w, r)
			}
		}))
		fmt.Println(err)
	}()
	//fmt.Println(proxyAddr)
	//proxy := goproxy.NewProxyHttpServer()
	//proxy.Tr.Proxy = http.ProxyURL(proxyAddr)
	//proxy.Verbose = true
	//
	//go func() {
	//	log.Fatal(http.ListenAndServe(":9090", proxy))
	//}()

	return "localhost:9090"
}
