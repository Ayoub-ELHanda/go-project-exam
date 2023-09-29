package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	stopScanningMutex sync.Mutex
	stopScanning      = false
)

func scanPort(ip string, port int, wg *sync.WaitGroup) {
	defer wg.Done()

	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, 1*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	fmt.Printf("Port %d is open\n", port)

	stopScanningMutex.Lock()
	if !stopScanning {
		// If the port is open, send a GET request to register the username
		respGet, err := http.Get(fmt.Sprintf("http://%s:%d/ping?username=ayoub", ip, port))
		if err != nil {
			fmt.Printf("Error making GET request to port %d: %v\n", port, err)
		} else {
			respGet.Body.Close()
			if respGet.StatusCode == http.StatusOK {
				fmt.Printf("Successfully registered username 'ayoub' on port %d using GET\n", port)
				stopScanning = true
			}
		}

		// If the port is open, send a POST request to register the username
		respPost, err := http.Post(fmt.Sprintf("http://%s:%d/signup", ip, port), "application/json", strings.NewReader(`{"USER":"ayoub"}`))
		if err != nil {
			fmt.Printf("Error making POST request to port %d: %v\n", port, err)
		} else {
			respPost.Body.Close()
			if respPost.StatusCode == http.StatusOK {
				fmt.Printf("Successfully registered username 'ayoub' on port %d using POST\n", port)
				stopScanning = true
			}
		}
	}
	stopScanningMutex.Unlock()
}

func main() {
	ip := "10.49.122.144"
	var wg sync.WaitGroup

	// Scan ports in the range from 1024 to 65535
	for port := 1024; port <= 65535; port++ {
		wg.Add(1)
		go scanPort(ip, port, &wg)

		stopScanningMutex.Lock()
		if stopScanning {
			stopScanningMutex.Unlock()
			break
		}
		stopScanningMutex.Unlock()
	}

	wg.Wait()
}
