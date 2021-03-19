package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	var startPort, endPort, workers uint = 1, 65535, 10
	var timeout = 60 * time.Second

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: portscanner [OPTION]... HOST\n\nOptions:\n")
		flag.CommandLine.PrintDefaults()
	}

	flag.UintVar(&startPort, "start", startPort, "start port")
	flag.UintVar(&endPort, "end", endPort, "end port")
	flag.UintVar(&workers, "workers", workers, "No. of workers")
	flag.DurationVar(&timeout, "timeout", timeout, "Timeout")

	flag.Parse()

	if flag.NArg() != 1 || startPort > 65535 || endPort > 65535 || startPort > endPort {
		flag.Usage()
		os.Exit(2)
	}

	host := flag.Arg(0)
	totalPorts := endPort + 1 - startPort
	ports := make(chan uint, totalPorts)
	for i := startPort; i <= endPort; i++ {
		ports <- i
	}
	close(ports)

	bar := pb.StartNew(int(totalPorts))

	results := make(chan uint, 100)

	var wg sync.WaitGroup
	wg.Add(int(workers))

	for i := uint(0); i < workers; i++ {
		go func() {
			for p := range ports {
				if dial(host, p, timeout) {
					results <- p
				}
				bar.Increment()
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Println(r)
	}
	bar.Finish()
}

func dial(host string, port uint, timeout time.Duration) bool {
	c, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))), timeout)
	if err != nil {
		return false
	}
	c.Close()
	return true
}
