package main

import "testing"

func TestBadServer(t *testing.T) {
	if serverNameValid("-1;example.com") {
		t.Error("Invalid characters passed!")
	}
}

func TestGoodServer(t *testing.T) {
	if !serverNameValid("srv.example.com") {
		t.Error("Valid server name failed!")
	}
}

func TestServerIP(t *testing.T) {
	if serverNameValid("127.0.0.1") {
		t.Error("FQDN passed!")
	}
}
