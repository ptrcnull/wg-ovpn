package main

import (
	"fmt"
	"golang.org/x/sys/unix"
)

const BufSize = 8192

func connect(one *pts, two *pts) error {
	_, err := unix.Open(one.Path(), unix.O_RDWR|unix.O_NOCTTY, 0777)
	if err != nil {
		return err
	}
	_, err = unix.Open(two.Path(), unix.O_RDWR|unix.O_NOCTTY, 0777)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, BufSize)
		for {
			n, err := unix.Read(one.ptmx, buf)
			if err != nil {
				panic(err)
			}

			//fmt.Println("one -> two", n)
			//fmt.Println(hex.Dump(buf[:n]))

			n, err = unix.Write(two.ptmx, buf[:n])
			if err != nil {
				panic(err)
			}

			fmt.Println("written", n)
		}
	}()

	go func() {
		buf := make([]byte, BufSize)
		for {
			n, err := unix.Read(two.ptmx, buf)
			if err != nil {
				panic(err)
			}

			//fmt.Println("two -> one", n)
			//fmt.Println(hex.Dump(buf[:n]))

			n, err = unix.Write(one.ptmx, buf[:n])
			if err != nil {
				panic(err)
			}

			fmt.Println("written", n)
		}
	}()

	return nil
}
