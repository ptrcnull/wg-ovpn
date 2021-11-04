package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func memfdCreate(filename string, content []byte) string {
	fd, err := unix.MemfdCreate(filename, 0)
	if err != nil {
		panic(fmt.Errorf("memfd create: %w", err))
	}

	err = unix.Ftruncate(fd, int64(len(content)))
	if err != nil {
		panic(fmt.Errorf("memfd truncate: %w", err))
	}

	_, err = unix.Write(fd, content)
	if err != nil {
		panic(fmt.Errorf("memfd write: %w", err))
	}

	return fmt.Sprintf("/proc/self/fd/%d", fd)
}
