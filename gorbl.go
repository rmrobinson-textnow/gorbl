/*
Package gorbl lets you perform RBL (Real-time Blackhole List - https://en.wikipedia.org/wiki/DNSBL)
lookups using Golang

This package takes inspiration from a similar module that I wrote in Python
(https://github.com/polera/rblwatch).

gorbl takes a simpler approach:  Basic lookup capability is provided by the
lib.  Unlike in rblwatch, concurrent lookups and the lists to search are
left to those using the lib.

JSON annotations on the types are provided as a convenience.
*/
package gorbl

import (
	"fmt"
	"net"
	"strings"

	"golang.org/x/net/context"
)

/*
RBL contains the lookup parameters for this blacklist.
*/
type RBL struct {
	// hostname is the DNS base to perform lookups against.
	hostname string
	// lookupTXT dictates whether we will also perform a TXT lookup for this blacklist.
	lookupTxt bool

	// resolver is an internal DNS resolver we will use (allowing for context to be passed to DNS lookups).
	resolver *net.Resolver
}

/*
RBLResults holds the results of the lookup.
*/
type RBLResults struct {
	// List is the RBL that was searched
	List string `json:"list"`
	// Host is the host or IP that was passed (i.e. smtp.gmail.com)
	Host string `json:"host"`
	// Results is a slice of Results - one per IP address searched
	Results []Result `json:"results"`
}

/*
Result holds the individual IP lookup results for each RBL search
*/
type Result struct {
	// Address is the IP address that was searched
	Address string `json:"address"`
	// Listed indicates whether or not the IP was on the RBL
	Listed bool `json:"listed"`
	// If the IP was listed, what address was returned?
	// RBL lists sometimes use the returned IP to indicate why it was listed.
	ListedAddress string `json:"listed_address"`
	// RBL lists sometimes add extra information as a TXT record
	// if any info is present, it will be stored here.
	Text string `json:"text"`
	// Error represents any error that was encountered (DNS timeout, host not
	// found, etc.) if any
	Error bool `json:"error"`
	// ErrorType is the type of error encountered if any
	ErrorType error `json:"error_type"`
}

// NewRBL creates a new RBL struct with the specified hostname and TXT lookup behaviour.
func NewRBL(hostname string, lookupTxt bool) *RBL {
	return &RBL{
		hostname:  hostname,
		lookupTxt: lookupTxt,
		resolver:  &net.Resolver{},
	}
}

/*
Reverse the octets of a given IPv4 address
64.233.171.108 becomes 108.171.233.64
*/
func Reverse(ip net.IP) string {
	if ip.To4() != nil {
		splitAddress := strings.Split(ip.String(), ".")

		for i, j := 0, len(splitAddress)-1; i < len(splitAddress)/2; i, j = i+1, j-1 {
			splitAddress[i], splitAddress[j] = splitAddress[j], splitAddress[i]
		}

		return strings.Join(splitAddress, ".")
	}
	return ""
}

/**
LookupIP looks up the specified IP in the RBL and returns its response.
 */
func (r *RBL) LookupIP(ctx context.Context, ip net.IP) RBLResults {
	ret := RBLResults{
		Host:    ip.String(),
		List:    r.hostname,
		Results: []Result{},
	}

	ipHostname := fmt.Sprintf("%s.%s", Reverse(ip), r.hostname)

	addrs, err := r.resolver.LookupHost(ctx, ipHostname)

	if len(addrs) < 1 {
		res := Result{
			Address: ip.String(),
			Listed:  false,
		}

		if err != nil {
			res.Error = true
			res.ErrorType = err
		}

		ret.Results = append(ret.Results, res)
		return ret
	}

	// For every IP address we get back the RBL IP lookup, we perform an optional TXT lookup.
	for _, addr := range addrs {
		res := Result{
			Address:       ip.String(),
			Listed:        true,
			ListedAddress: addr,
		}

		if r.lookupTxt {
			txt, _ := r.resolver.LookupTXT(ctx, ipHostname)

			// We skip both empty results and errors.
			if len(txt) > 0 {
				res.Text = txt[0]
			}
		}

		if err != nil {
			res.Error = true
			res.ErrorType = err
		}

		ret.Results = append(ret.Results, res)
	}

	return ret
}

/*
Lookup performs a search for IPs tied to the specified hostname and returns the response.
*/
func (r *RBL) Lookup(ctx context.Context, targetHost string) RBLResults {
	ret := RBLResults{
		Host:    targetHost,
		List:    r.hostname,
		Results: []Result{},
	}

	// Find all IP addresses associated with the supplied hostname.
	if addrs, err := r.resolver.LookupIPAddr(ctx, targetHost); err == nil {
		for _, addr := range addrs {
			// For every valid IPv4 address tied to this hostname, we perform an RBL lookup.
			if addr.IP.To4() != nil {
				qResults := r.LookupIP(ctx, addr.IP)

				ret.Results = append(ret.Results, qResults.Results...)
			}
		}
	}

	return ret
}
