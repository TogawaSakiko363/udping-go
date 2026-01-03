package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultInterval = 1000
	defaultLen      = 64
)

var (
	targetIP   string
	targetPort int
	payloadLen int
	interval   int
	isIPv6     bool
)

var (
	count         int
	countReceived int
	rttSum        float64
	rttMin        float64
	rttMax        float64
)

func init() {
	rttMin = 99999999.0
	flag.StringVar(&targetIP, "ip", "", "Target IP address")
	flag.IntVar(&targetPort, "port", 0, "Target port")
	flag.IntVar(&payloadLen, "len", defaultLen, "Payload length in bytes")
	flag.IntVar(&interval, "interval", defaultInterval, "Interval in milliseconds")
	flag.BoolVar(&isIPv6, "6", false, "Use IPv6")
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func printStatistics() {
	fmt.Println()
	fmt.Println("--- ping statistics ---")
	if count > 0 {
		lossPercent := float64(count-countReceived) * 100.0 / float64(count)
		fmt.Printf("%d packets transmitted, %d received, %.2f%% packet loss\n", count, countReceived, lossPercent)
	}
	if countReceived > 0 {
		avgRtt := rttSum / float64(countReceived)
		fmt.Printf("rtt min/avg/max = %.2f/%.2f/%.2f ms\n", rttMin, avgRtt, rttMax)
	}
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	printStatistics()
	os.Exit(0)
}

func createConnection() (*net.UDPConn, error) {
	if isIPv6 {
		return net.DialUDP("udp6", nil, &net.UDPAddr{
			IP:   net.ParseIP(targetIP),
			Port: targetPort,
		})
	}
	return net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP(targetIP),
		Port: targetPort,
	})
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage:")
		fmt.Println("  udpping <dest_ip> <dest_port>")
		fmt.Println("  udpping <dest_ip> <dest_port> [options]")
		fmt.Println()
		fmt.Println("options:")
		fmt.Println("  -len        Payload length in bytes (default: 64)")
		fmt.Println("  -interval   Interval between packets in milliseconds (default: 1000)")
		fmt.Println("  -6          Use IPv6")
		fmt.Println()
		fmt.Println("examples:")
		fmt.Println("  udpping 44.55.66.77 4000")
		fmt.Println("  udpping fe80::5400:ff:aabb:ccdd 4000 -6")
		fmt.Println("  udpping 44.55.66.77 4000 -len=400 -interval=2000")
	}

	flag.Parse()

	if targetIP == "" || targetPort == 0 {
		if len(os.Args) < 3 {
			flag.Usage()
			os.Exit(1)
		}

		remainingArgs := flag.Args()
		if len(remainingArgs) < 2 {
			flag.Usage()
			os.Exit(1)
		}

		targetIP = remainingArgs[0]
		fmt.Sscanf(remainingArgs[1], "%d", &targetPort)

		for i := 2; i < len(remainingArgs); i++ {
			arg := remainingArgs[i]
			if strings.HasPrefix(arg, "-len=") {
				fmt.Sscanf(arg, "-len=%d", &payloadLen)
			} else if strings.HasPrefix(arg, "-interval=") {
				fmt.Sscanf(arg, "-interval=%d", &interval)
			} else if arg == "-6" {
				isIPv6 = true
			}
		}
	}

	if strings.Contains(targetIP, ":") {
		isIPv6 = true
	}

	if payloadLen < 5 {
		fmt.Println("LEN must be >=5")
		os.Exit(1)
	}

	if interval < 50 {
		fmt.Println("INTERVAL must be >=50")
		os.Exit(1)
	}

	go signalHandler()

	fmt.Printf("UDPping %s via port %d with %d bytes of payload\n", targetIP, targetPort, payloadLen)

	intervalDuration := time.Duration(interval) * time.Millisecond

	var conn *net.UDPConn
	var err error

	conn, err = createConnection()
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		os.Exit(1)
	}

	for {
		payload := randomString(payloadLen)
		timeOfSend := time.Now()

		_, err := conn.Write([]byte(payload))
		if err != nil {
			conn.Close()
			conn, err = createConnection()
			if err != nil {
				fmt.Printf("Error reconnecting: %v\n", err)
				time.Sleep(intervalDuration)
				continue
			}
			_, err = conn.Write([]byte(payload))
			if err != nil {
				fmt.Printf("Error sending: %v\n", err)
				time.Sleep(intervalDuration)
				continue
			}
		}

		deadline := timeOfSend.Add(intervalDuration)
		conn.SetDeadline(deadline)

		received := false
		rtt := 0.0

		buf := make([]byte, 65536)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					break
				}
				fmt.Printf("Error reading: %v\n", err)
				break
			}

			receivedData := string(buf[:n])
			receivedIP := addr.IP.String()

			if receivedData == payload && receivedIP == targetIP && addr.Port == targetPort {
				rtt = float64(time.Since(timeOfSend).Microseconds()) / 1000.0
				fmt.Printf("Reply from %s seq=%d time=%.2f ms\n", targetIP, count, rtt)
				received = true
				break
			}
		}

		count++

		if received {
			countReceived++
			rttSum += rtt
			if rtt > rttMax {
				rttMax = rtt
			}
			if rtt < rttMin {
				rttMin = rtt
			}
		} else {
			fmt.Println("Request timed out")
		}

		timeRemaining := time.Until(deadline)
		if timeRemaining > 0 {
			time.Sleep(timeRemaining)
		}
	}
}
