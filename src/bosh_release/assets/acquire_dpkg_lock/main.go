package main

import (
	"io"
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	name := "/var/lib/dpkg/lock"
	file, err := os.OpenFile(name, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
	if err != nil {
		log.Printf("error opening file: %s", err)
		return
	}
	defer file.Close()

	flockT := syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: io.SeekStart,
		Start:  0,
		Len:    0,
	}
	err = syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, &flockT)
	if err != nil {
		log.Printf("error locking file: %s", err)
		return
	}

	log.Println("locked " + name)

	time.Sleep(5 * time.Hour)
}
