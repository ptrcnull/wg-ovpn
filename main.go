package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
)

//go:embed openvpn
var openvpn []byte

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: wg-ovpn <.ovpn file> <wireguard config file>")
		os.Exit(1)
	}

	ptsWg, err := allocatePts()
	check(err)
	defer ptsWg.Close()

	ptsOvpn, err := allocatePts()
	check(err)
	defer ptsOvpn.Close()

	err = connect(ptsWg, ptsOvpn)
	check(err)

	wgConfig, err := ioutil.ReadFile(os.Args[2])
	check(err)

	config, err := ioutil.ReadFile(os.Args[1])
	check(err)

	sconfig := string(config)

	devNode := "dev-node "+ptsOvpn.Path()
	if strings.Contains(sconfig, "dev-node ") {
		sconfig = regexp.MustCompile("dev-node .*").ReplaceAllString(sconfig, devNode)
	} else {
		sconfig = devNode + "\n" + sconfig
	}

	up := "up \"/bin/sh -c 'echo $@ >/dev/stderr'\""
	if strings.Contains(sconfig, "\nup ") {
		sconfig = regexp.MustCompile("^up .*").ReplaceAllString(sconfig, up)
	} else {
		sconfig = up + "\n" + sconfig
	}

	sconfig = "route-noexec\nifconfig-noexec\n" + sconfig
	//log.Println("final config:")
	//fmt.Println(sconfig)
	config = []byte(sconfig)

	ovpnExec := memfdCreate("openvpn", openvpn)
	ovpnConfig := memfdCreate("config.ovpn", config)
	ovpnOutput := bytes.NewBuffer([]byte{})
	ovpnCmd := exec.Command(ovpnExec, ovpnConfig)
	ovpnCmd.Stdout = os.Stdout
	ovpnCmd.Stderr = ovpnOutput

	go func() {
		err := ovpnCmd.Run()
		log.Println("openvpn exited")
		if err != nil {
			log.Println(err)
		}
		os.Exit(1)
	}()

	file, err := os.OpenFile(ptsWg.Path(), os.O_RDWR|unix.O_NOCTTY, 0777)
	check(err)
	tun := &FakeTun{file: file}

	for ovpnOutput.Len() == 0 {
	}
	ovpnData := strings.Split(ovpnOutput.String(), " ")
	tun.clientIp = tcpip.Address(net.ParseIP(ovpnData[2]).To4())

	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelVerbose, ""))
	dev.IpcSet(string(wgConfig))
	err = dev.Up()
	check(err)

	select {}
}
