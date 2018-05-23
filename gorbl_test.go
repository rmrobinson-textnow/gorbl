package gorbl

import (
	"net"
	"testing"

	"golang.org/x/net/context"
)

func TestReverseIP(t *testing.T) {
	t.Parallel()
	ip := net.IP{192, 168, 1, 1}

	r := Reverse(ip)

	if r != "1.1.168.192" {
		t.Errorf("Expected ip to equal 1.1.168.192, actual %s", r)
	}
}

func TestLookupParams(t *testing.T) {
	t.Parallel()
	rblName := "b.barracudacentral.org"
	rbl := NewRBL(rblName, true)

	res := rbl.Lookup(context.Background(), "smtp.gmail.com")

	if res.List != rblName {
		t.Errorf("Expected b.barracudacentral.org, actual %s", res.List)
	}

	if res.Host != "smtp.gmail.com" {
		t.Errorf("Expected smtp.gmail.com, actual %s", res.Host)
	}
}
