package main

import (
	"encoding/binary"
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
	log.Println("Read")
	i, err := f.file.Read(bytes[offset:])
	log.Println("READ", i, err)
	if err != nil {
		return 0, err
	}

	packetInfo(bytes[offset:])

	hdr := header.IPv4(bytes[offset:offset+20])
	hdr.SetDestinationAddressWithChecksumUpdate(f.sourceIp)

	packetInfo(bytes[offset:])

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
	log.Println("Write")
	bytes = bytes[offset:]

	packetInfo(bytes)

	f.sourceIp = tcpip.Address(net.IPv4(bytes[12], bytes[13], bytes[14], bytes[15]).To4())
	hdr := header.IPv4(bytes[:20]) // nobody uses options anyway ...right?
	hdr.SetSourceAddressWithChecksumUpdate(f.clientIp)

	packetInfo(bytes)

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

// Checksum is the "internet checksum" from https://tools.ietf.org/html/rfc1071.
func checksum(buf []byte, initial uint16) []byte {
	v := uint32(initial)
	for i := 0; i < len(buf)-1; i += 2 {
		v += uint32(binary.BigEndian.Uint16(buf[i:]))
	}
	if len(buf)%2 == 1 {
		v += uint32(buf[len(buf)-1]) << 8
	}
	for v > 0xffff {
		v = (v >> 16) + (v & 0xffff)
	}
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, ^uint16(v))
	return res
}
