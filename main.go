package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type FileEventHandler func(fd int)

type FileEventLoop struct {
	fd        int
	epollFd   int
	eventType uint32
	handler   FileEventHandler
}

func NewFileEventLoop(filePath string, eventType uint32, handler FileEventHandler) (*FileEventLoop, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fd := int(file.Fd())

	epollFd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	event := unix.EpollEvent{
		Events: eventType,
		Fd:     int32(fd),
	}

	err = unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		return nil, err
	}

	return &FileEventLoop{
		fd:        fd,
		epollFd:   epollFd,
		eventType: eventType,
		handler:   handler,
	}, nil
}

func (f *FileEventLoop) Run() {
	events := make([]unix.EpollEvent, 1)

	for {
		numEvents, err := unix.EpollWait(f.epollFd, events, -1)
		if err != nil {
			fmt.Println("Error waiting for events:", err)
			return
		}

		for i := 0; i < numEvents; i++ {
			if events[i].Fd == int32(f.fd) {
				f.handler(f.fd)
			}
		}
	}
}

func main() {
	fileHandler := func(fd int) {
		fmt.Println("Handling file with descriptor:", fd)
		// Your file handling logic goes here
	}

	fileEventLoop, err := NewFileEventLoop("example.txt", unix.EPOLLIN, fileHandler)
	if err != nil {
		fmt.Println("Error creating file event loop:", err)
		return
	}
	defer unix.Close(fileEventLoop.epollFd)

	fileEventLoop.Run()
}
