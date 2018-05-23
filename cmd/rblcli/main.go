package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/rmrobinson-textnow/gorbl"
	"golang.org/x/net/context"
)

func main() {
	var (
		host = flag.String("host", "", "The host to lookup. Mutually exclusive to IP")
		ip   = flag.String("ip", "", "The IP to lookup. Mutually exclusive to host")
	)

	flag.Parse()

	if len(*host) > 0 && len(*ip) > 0 {
		fmt.Printf("Only one of IP or host can be supplied\n")
		return
	}

	rbl := gorbl.NewRBL("bl.mailspike.net", true)

	var ret gorbl.RBLResults

	if len(*host) > 0 {
		ret = rbl.Lookup(context.Background(), *host)
	} else if len(*ip) > 0 {
		parsedIP := net.ParseIP(*ip)

		if parsedIP == nil {
			fmt.Printf("Supplied IP unable to be parsed\n")
			return
		}

		ret = rbl.LookupIP(context.Background(), parsedIP)
	} else {
		fmt.Printf("One of IP or host must be supplied\n")
		return
	}

	for _, res := range ret.Results {
		fmt.Printf("%+v\n", res)
	}
}
