package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type windData struct {
	stow, batt       int
	head, speed, avg float32
}

func parseWindData(pack [17]byte) windData {

}

func main() {

	ping := []byte("PING")

	serveraddr, err := net.ResolveUDPAddr("udp", "windyg.phys.utas.edu.au:7758")
	if err != nil {
		log.Fatal(err)
	}

	localaddr, err := net.ResolveUDPAddr("udp", ":7755")
	if err != nil {
		log.Fatal(err)
	}

	Conn, err := net.DialUDP("udp", localaddr, serveraddr)
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()

	buf := make([]byte, 1024)
	i := 0

	for {
		i++
		_, err := Conn.Write(ping)
		log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "PING")
		if err != nil {
			log.Fatal(err)
		}

		err = Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		n, addr, err := Conn.ReadFromUDP(buf)

		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			log.Println("pong timed out")
			continue
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s:%d - %s\n", addr.IP.String(), addr.Port, string(buf[0:n]))

		fmt.Println("")
		time.Sleep(time.Second * 1)
	}

}
