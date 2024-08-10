package znet

import (
	"net"
	"testing"
)

func TestNewWithLocalAddr(t *testing.T) {
	var u = "https://api.ipify.org"
	var ip = "192.168.0.1"

	client := NewWithLocalAddr(&net.TCPAddr{IP: net.ParseIP(ip)})
	resp, err := client.R().SetQueryParam("format", "json").Get(u)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(resp.String())
}

func TestBackoff(t *testing.T) {
	var u = "https://cc.ipify.org?format=json"

	client := New()
	resp, err := client.R().GetWithRetries(u)
	//resp, err := client.R().GetWithRetries(u, Retries(1), WaitTime(5*time.Second))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(resp.String())
}
