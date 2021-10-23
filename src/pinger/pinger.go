package pinger

import (
	"os"
	"time"
	"net"
	"strings"
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// TODO: Add linting
// TODO: Add argument parsing for `n` requests

type ipVersion int
type ipNumber int

const (
	IPV4 ipVersion = iota
	IPV6 ipVersion = iota
)

var protocolToNumber = map[ipVersion]int {
	IPV4: 1,
	IPV6: 58,
}

type EchoReply struct {
	n int
	peer net.Addr
	msg *icmp.Message
}

func ResolveHost(host string, ip ipVersion) (string, error){
	addrs, err := net.LookupHost(host)
	var addr string
	if err != nil {
		return "", err
	}
	switch ip {
	case IPV6:
		addr = addrs[0]
	case IPV4:
		addr = addrs[1]
	default:
		log.Fatal("Invalid IP version detected on resolve.")
	}
	return addr, nil
}

func InitICMPListen(ip ipVersion, timeout time.Duration) (*icmp.PacketConn, error) {
	var listenIPType string
	var listenAddr string

	switch ip {
	case IPV6:
		listenIPType, listenAddr = "ip6:ipv6-icmp", "::"
	case IPV4:
		listenIPType, listenAddr = "ip4:icmp", "0.0.0.0"
	default:
		return nil, errors.New("Invalid IP version detected.")
	}

	conn, err := icmp.ListenPacket(listenIPType, listenAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func BuildPacket(ip ipVersion, seqNum int) (icmp.Message) {
	var proto icmp.Type
	switch ip {
	case IPV6:
		proto = ipv6.ICMPTypeEchoRequest
	case IPV4:
		proto = ipv4.ICMPTypeEcho
	}

	packet := icmp.Message {
		Type: proto, 
		Code: 0,
		Body: &icmp.Echo {
			ID: os.Getpid(), 
			Seq: seqNum,
			Data: []byte("moshi moshi moshi moshi!"),
		},
	}

	return packet
}

func Receive(conn *icmp.PacketConn, protocol ipVersion) (EchoReply, error){
	reply := new(EchoReply)
	buffer := make([]byte, 66507)
    n, peer, err := conn.ReadFrom(buffer)
    if err != nil {
		return *reply, err
    }

    msg, err := icmp.ParseMessage(protocolToNumber[protocol], buffer[:n])
    if err != nil {
        log.Fatal(err)
    }
	reply.n = n
	reply.msg = msg
	reply.peer = peer

	return *reply, nil
}

func ping(host string, protocol ipVersion) {
	timeout := time.Second / 1000 * 100
	log.SetLevel(log.InfoLevel)

	addr, err := ResolveHost(host, protocol)
	if err != nil {
		log.Fatal("Could not resolve hostname.", err)
	}
	conn, err := InitICMPListen(protocol, timeout)
	if err != nil {
		log.Fatal("Could not establish listener.", err)
	}
	defer conn.Close()

	packetNum := 1
	for {
		packet := BuildPacket(protocol, packetNum)
		packetM, err := packet.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}

		// Send ICMP packet
		_, err = conn.WriteTo(packetM, &net.IPAddr{IP: net.ParseIP(addr)})
		if err != nil {
			log.Fatalf("WriteTo err, %s", err)
		}

		// Begin latency recording
		sendTime := time.Now()
		// Receive ICMP echo
		err = conn.SetReadDeadline(time.Now().Add(timeout))
		reply, err := Receive(conn, protocol)
		recvTime := time.Since(sendTime).Round(time.Millisecond)

		// Register request timed out
		if err != nil && strings.Contains(err.Error(), "timeout") {
			log.Info("Request timed out.")
			time.Sleep(1 * time.Second)
			continue
		} 
		if err != nil {
			log.Fatal(err)
		}

		switch reply.msg.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			log.Printf("Reply from %v: bytes=%d time=%s", reply.peer, reply.n, recvTime)
		default:
			log.Warn("Ignoring non-ICMP reply.")
		}
		packetNum++
		time.Sleep(1 * time.Second)
	}
}
