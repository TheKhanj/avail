package main

import (
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

func NewProxiedHttpClient(addr string) (*http.Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		t := &http.Transport{
			Proxy: http.ProxyURL(u),
		}
		return &http.Client{Transport: t}, nil

	case "socks5":
		var auth *proxy.Auth
		if u.User != nil {
			password, _ := u.User.Password()
			auth = &proxy.Auth{
				User:     u.User.Username(),
				Password: password,
			}
		}

		host := u.Host
		dialer, err := proxy.SOCKS5("tcp", host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		t := &http.Transport{
			Dial: dialer.Dial,
		}
		return &http.Client{Transport: t}, nil

	default:
		return nil, errors.New("unsupported proxy scheme")
	}
}
