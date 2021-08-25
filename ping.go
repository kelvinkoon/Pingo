package main

import (
	"fmt"
	"net"
)

// func buildICMPPacket()
// TODO: Add logrus
// TODO: Add linting

func ping(addr string) {
	fmt.Println("Pinging...")
	// Resolve IP address
	ip, err := net.LookupHost(addr)
	if err != nil {
		fmt.Println("Lookup failed")
	}
	fmt.Println(ip)
	// Initialize ICMP reply socket
	// Build ICMP packet
	// Initialize ICMP sending socket
	// Send ICMP packet
	// Read from reply socket
	return
}

func main() {
    fmt.Println("Hello, World!")
	ping("google.com")
}
