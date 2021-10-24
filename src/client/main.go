package main

import (
	"flag"
	p "github.com/kelvinkoon/Pingo/src/pinger"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Parse arguments
	hostPtr := flag.String("host", "github.com", "a string")
	protoPtr := flag.String("ip", "4", "a string")
	flag.Parse()

	proto := p.UNKNOWN
	switch *protoPtr {
	case "4":
		proto = p.IPV4
	case "6":
		proto = p.IPV6
	default:
		log.Fatal("Please provide '4' or '6' for IP version (eg. `-ip=4`")
	}

	p.Ping(*hostPtr, proto)
}
