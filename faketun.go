package main

import (
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"io"
	"log"
	"net"
	"os"

	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

type FakeTun struct {
	file     io.ReadWriteCloser
	clientIp tcpip.Address
	sourceIp tcpip.Address
}

func (f *FakeTun) File() *os.File {
	return f.file.(*os.File)
}

func (f *FakeTun) Read(bytes []byte, offset int) (int, error) {
	//log.Println("Read")
	i, err := f.file.Read(bytes[offset:])
	if err != nil {
		return 0, err
	}

	//packetInfo(bytes[offset:])

	hdr := header.IPv4(bytes[offset:offset+header.IPv4MaximumHeaderSize])
	hdr.SetDestinationAddressWithChecksumUpdate(f.sourceIp)

	//packetInfo(bytes[offset:])

	return i, nil
}

func packetInfo(bytes []byte) {
	log.Printf(
		"proto %d checksum %x from %s to %s\n",
		bytes[9], bytes[10:12],
		net.IPv4(bytes[12], bytes[13], bytes[14], bytes[15]),
		net.IPv4(bytes[16], bytes[17], bytes[18], bytes[19]),
	)
}

func (f *FakeTun) Write(bytes []byte, offset int) (int, error) {
	//log.Println("Write")
	bytes = bytes[offset:]

	hdr := header.IPv4(bytes[:header.IPv4MinimumSize])
	// nobody uses options anyway ...right?
	f.sourceIp = hdr.SourceAddress()

	//packetInfo(bytes)
	//log.Println(hdr.Protocol(), hdr.Checksum(), hdr.IsChecksumValid(), hdr.SourceAddress(), hdr.DestinationAddress())

	hdr.SetSourceAddressWithChecksumUpdate(f.clientIp)

	// fix tcp checksum
	if hdr.Protocol() == uint8(header.TCPProtocolNumber) {
		hdrTcp := header.TCP(bytes[header.IPv4MinimumSize:header.IPv4MinimumSize+header.TCPMinimumSize])
		hdrTcp.UpdateChecksumPseudoHeaderAddress(f.sourceIp, f.clientIp, true)
	}

	//packetInfo(bytes)
	//log.Println(hdr.Protocol(), hdr.Checksum(), hdr.IsChecksumValid(), hdr.SourceAddress(), hdr.DestinationAddress())

	//fmt.Println(hex.Dump(bytes))
	return f.file.Write(bytes)
}

func (f *FakeTun) Flush() error {
	return nil
}

func (f *FakeTun) MTU() (int, error) {
	return device.DefaultMTU, nil
}

func (f *FakeTun) Name() (string, error) {
	return "", nil
}

func (f *FakeTun) Events() chan tun.Event {
	ch := make(chan tun.Event)
	return ch
}

func (f *FakeTun) Close() error {
	return f.file.Close()
}
