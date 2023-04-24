package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type ArpResult struct {
	Ip  net.IP
	Mac net.HardwareAddr
}

func scan(interfaceName string, timeout time.Duration) ([]ArpResult, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var addr *net.IPNet
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				addr = &net.IPNet{
					IP:   ip4,
					Mask: ipnet.Mask[len(ipnet.Mask)-4:],
				}
				break
			}
		}
	}
	if addr == nil {
		return nil, errors.New("no good IP network found")
	} else if addr.IP[0] == 127 {
		return nil, errors.New("skipping localhost")
	} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return nil, errors.New("mask means network is too large")
	}

	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	result := make(chan ArpResult)
	defer close(result)

	stop := make(chan struct{})
	go readARP(handle, iface, stop, result)
	defer close(stop)
	// Write our scan packets out to the handle.
	if err := writeARP(handle, iface, addr); err != nil {
		// color.Yellow("error writing packets on %v: %v", iface.Name, err)
		return nil, err
	}
	// We don't know exactly how long it'll take for packets to be
	// sent back to us, but 10 seconds should be more than enough
	// time ;)

	found_ips := map[string]ArpResult{}
	for _, ip := range ips(addr) {
		found_ips[ip.String()] = ArpResult{Ip: ip, Mac: nil}
	}
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			vs := []ArpResult{}
			for _, v := range found_ips {
				vs = append(vs, v)
			}
			return vs, nil
		case packet := <-result:
			found_ips[packet.Ip.String()] = packet
		}
	}
}

func readARP(handle *pcap.Handle, iface *net.Interface, stop chan struct{}, result chan ArpResult) {
	w := tabwriter.NewWriter(os.Stdout, 20, 20, 1, ' ', tabwriter.DiscardEmptyColumns)
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		var packet gopacket.Packet
		select {
		case <-stop:
			return
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)
			if arp.Operation != layers.ARPReply || bytes.Equal([]byte(iface.HardwareAddr), arp.SourceHwAddress) {
				// This is a packet I sent.
				continue
			}
			// Note:  we might get some packets here that aren't responses to ones we've sent,
			// if for example someone else sends US an ARP request.  Doesn't much matter, though...
			// all information is good information :)

			fmt.Fprintf(w, "%v\t%v\n", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))
			w.Flush()
			ip := ArpResult{net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress)}
			result <- ip
		}
	}
}

func writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet) error {
	// Set up all the layers' fields we can.
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(addr.IP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}
	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// Send one packet for every address.
	for _, ip := range ips(addr) {
		arp.DstProtAddress = []byte(ip)
		gopacket.SerializeLayers(buf, opts, &eth, &arp)
		if err := handle.WritePacketData(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// ips is a simple and not very good method for getting all IPv4 addresses from a
// net.IPNet.  It returns all IPs it can over the channel it sends back, closing
// the channel when done.
func ips(n *net.IPNet) (out []net.IP) {
	num := binary.BigEndian.Uint32([]byte(n.IP))
	mask := binary.BigEndian.Uint32([]byte(n.Mask))
	num &= mask
	for mask < 0xffffffff {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], num)
		out = append(out, net.IP(buf[:]))
		mask++
		num++
	}
	return
}
