package pinger

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"os"
	"strings"
	"time"
)

type ipVersion int
type ipNumber int

const (
	IPV4    ipVersion = iota
	IPV6    ipVersion = iota
	UNKNOWN ipVersion = iota
)

var protocolToNumber = map[ipVersion]int{
	IPV4: 1,
	IPV6: 58,
}

type EchoReply struct {
	n    int
	peer net.Addr
	msg  *icmp.Message
}

func init() {
	log.SetLevel(log.InfoLevel)
}

func resolveHost(host string, ip ipVersion) (string, error) {
	var selectedAddr string
	var ipv4Addr string
	var ipv6Addr string

	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}

	// If single address given, assume IPv4 only
	if len(addrs) == 1 {
		ipv4Addr = addrs[0]
	} else {
		ipv4Addr = addrs[1]
		ipv6Addr = addrs[0]
	}

	switch ip {
	case IPV6:
		selectedAddr = ipv6Addr
	case IPV4:
		selectedAddr = ipv4Addr
	default:
		log.Fatal("Invalid IP version detected on resolve.")
	}
	return selectedAddr, nil
}

func initICMPListen(ip ipVersion, timeout time.Duration) (*icmp.PacketConn, error) {
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

func buildPacket(ip ipVersion, seqNum int) icmp.Message {
	var proto icmp.Type
	switch ip {
	case IPV6:
		proto = ipv6.ICMPTypeEchoRequest
	case IPV4:
		proto = ipv4.ICMPTypeEcho
	}

	packet := icmp.Message{
		Type: proto,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  seqNum,
			Data: []byte("moshi moshi moshi moshi!"),
		},
	}

	return packet
}

func receive(conn *icmp.PacketConn, protocol ipVersion) (EchoReply, error) {
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

func Ping(host string, protocol ipVersion) {
	timeout := time.Second / 1000 * 100

	addr, err := resolveHost(host, protocol)
	if err != nil {
		log.Fatal("Could not resolve hostname.", err)
	}
	if addr == "" {
		log.Fatal("Hostname could not be resolved, likely due to host supporting IPv4 only.")
	}

	conn, err := initICMPListen(protocol, timeout)
	if err != nil {
		log.Fatal("Could not establish listener.", err)
	}
	defer conn.Close()

	packetNum := 1
	for {
		packet := buildPacket(protocol, packetNum)
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
		reply, err := receive(conn, protocol)
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
			log.Debug("Ignoring non-ICMP reply.")
		}

		packetNum++
		time.Sleep(1 * time.Second)
	}
}
