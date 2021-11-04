package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"syscall"
)

type pts struct {
	fd   int
	ptmx int
}

func (p *pts) Open() (int, error) {
	return unix.Open(p.Path(), unix.O_RDWR|unix.O_NOCTTY|unix.O_LARGEFILE, 0777)
}

func (p *pts) Path() string {
	return fmt.Sprintf("/dev/pts/%d", p.fd)
}

func (p *pts) Close() {
	_ = unix.Close(p.ptmx)
}

func allocatePts() (*pts, error) {
	ptmx, err := unix.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open ptmx: %w", err)
	}

	res, err := unix.IoctlGetInt(ptmx, syscall.TIOCSPTLCK)
	if err != nil {
		return nil, fmt.Errorf("get ptmx lock: %w", err)
	}
	log.Println("ptmx lock status:", res)

	fd, err := unix.IoctlGetInt(ptmx, syscall.TIOCGPTN)
	if err != nil {
		return nil, fmt.Errorf("allocate pts: %w", err)
	}
	log.Println("allocated pts:", fd)

	_, err = unix.FcntlInt(uintptr(ptmx), unix.F_SETFD, unix.FD_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("ptmx setfd: %w", err)
	}

	term, err := unix.IoctlGetTermios(ptmx, unix.TCGETS)
	if err != nil {
		return nil, fmt.Errorf("tcgets: %w", err)
	}

	term.Iflag = 0
	term.Oflag ^= unix.OPOST
	term.Lflag ^= unix.ISIG | unix.ICANON | unix.ECHO

	err = unix.IoctlSetTermios(ptmx, unix.TCSETSW, term)
	if err != nil {
		return nil, fmt.Errorf("tcsetsw: %w", err)
	}

	return &pts{
		fd:   fd,
		ptmx: ptmx,
	}, nil
}
