package zeroconf

import (
	"fmt"
	"net"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	// Multicast groups used by mDNS
	mdnsGroupIPv4 = net.IPv4(224, 0, 0, 251)
	mdnsGroupIPv6 = net.ParseIP("ff02::fb")

	// mDNS wildcard addresses
	mdnsWildcardAddrIPv4 = &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.0"),
		Port: 5353,
	}
	mdnsWildcardAddrIPv6 = &net.UDPAddr{
		IP: net.ParseIP("ff02::"),
		// IP:   net.ParseIP("fd00::12d3:26e7:48db:e7d"),
		Port: 5353,
	}

	// mDNS endpoint addresses
	ipv4Addr = &net.UDPAddr{
		IP:   mdnsGroupIPv4,
		Port: 5353,
	}
	ipv6Addr = &net.UDPAddr{
		IP:   mdnsGroupIPv6,
		Port: 5353,
	}
)

func joinUdp6Multicast(interfaces []net.Interface) (map[int]*ipv6.PacketConn, error) {
	if len(interfaces) == 0 {
		interfaces = listMulticastInterfaces()
	}
	// log.Println("Using multicast interfaces: ", interfaces)

	var pkConns = map[int]*ipv6.PacketConn{}
	var failedJoins int
	for _, iface := range interfaces {
		udpConn, err := net.ListenMulticastUDP("udp6", &iface, mdnsWildcardAddrIPv6)
		if err != nil {
			failedJoins++
			continue
		}

		// Join multicast groups to receive announcements
		pkConn := ipv6.NewPacketConn(udpConn)
		pkConn.SetControlMessage(ipv6.FlagInterface, true)

		if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroupIPv6}); err != nil {
			// log.Println("Udp6 JoinGroup failed for iface ", iface)
			pkConn.Close()
			failedJoins++
			continue
		}
		pkConns[iface.Index] = pkConn
	}
	if failedJoins == len(interfaces) {
		//pkConn.Close()
		return nil, fmt.Errorf("udp6: failed to join any of these interfaces: %v", interfaces)
	}

	return pkConns, nil
}

func joinUdp4Multicast(interfaces []net.Interface) (map[int]*ipv4.PacketConn, error) {
	if len(interfaces) == 0 {
		interfaces = listMulticastInterfaces()
	}
	// log.Println("Using multicast interfaces: ", interfaces)

	var pkConns = map[int]*ipv4.PacketConn{}
	var failedJoins int
	for _, iface := range interfaces {
		udpConn, err := net.ListenMulticastUDP("udp4", &iface, mdnsWildcardAddrIPv4)
		if err != nil {
			// log.Printf("[ERR] bonjour: Failed to bind to udp4 mutlicast: %v", err)
			failedJoins++
			continue
		}

		// Join multicast groups to receive announcements
		pkConn := ipv4.NewPacketConn(udpConn)
		pkConn.SetControlMessage(ipv4.FlagInterface, true)

		if err := pkConn.JoinGroup(&iface, &net.UDPAddr{IP: mdnsGroupIPv4}); err != nil {
			// log.Println("Udp4 JoinGroup failed for iface ", iface)
			pkConn.Close()
			failedJoins++
			continue
		}

		pkConns[iface.Index] = pkConn
	}
	if failedJoins == len(interfaces) {
		//pkConn.Close()
		return nil, fmt.Errorf("udp4: failed to join any of these interfaces: %v", interfaces)
	}

	return pkConns, nil
}

func listMulticastInterfaces() []net.Interface {
	var interfaces []net.Interface
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ifaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		if (ifi.Flags & net.FlagMulticast) > 0 {
			interfaces = append(interfaces, ifi)
		}
	}

	return interfaces
}
