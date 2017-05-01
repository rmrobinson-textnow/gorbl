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
)

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

func query(rbl string, ip net.IP, lookupTxt bool) []Result {
    ret := []Result{}

	lookup := fmt.Sprintf("%s.%s", Reverse(ip), rbl)

	addrs, err := net.LookupHost(lookup)

    if len(addrs) < 1 {
        ret = append(ret, Result{
            Address: ip.String(),
            Listed: false,
        })

        if err != nil {
            ret[0].Error = true
            ret[0].ErrorType = err
        }

        return ret
    }

    for _, addr := range addrs {
        r := Result{
            Address: ip.String(),
            Listed: true,
            ListedAddress: addr,
        }

        if lookupTxt {
            txt, _ := net.LookupTXT(lookup)

            if len(txt) > 0 {
                r.Text = txt[0]
            }
        }

        if err != nil {
            r.Error = true
            r.ErrorType = err
        }

        ret = append(ret, r)
    }

	return ret
}

/*
Lookup performs the search and returns the RBLResults
*/
func Lookup(rblList string, targetHost string, lookupTxt bool) (r RBLResults) {
	r.List = rblList
	r.Host = targetHost

	if ip, err := net.LookupIP(targetHost); err == nil {
		for _, addr := range ip {
			if addr.To4() != nil {
                qResults := query(rblList, addr, lookupTxt)

				r.Results = append(r.Results, qResults...)
			}
		}
	}
	return
}
