package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type windData struct {
	stow, batt       int
	head, speed, avg float64
}

func parseWindData(pack []byte) (w windData, err error) {
	if pack[0] == 1 {
		err = errors.New("Not valid packet")
	}

	npack := make([]byte, 17)
	copy(npack, pack[0:16])
	for i := 0; i < 14; i++ {
		npack[i] += '0'
	}

	w.stow = int(npack[15])
	w.batt = int(npack[14])

	w.head, err = strconv.ParseFloat(string(npack[1:4]), 32)
	if err != nil {
		return
	}

	w.speed, err = strconv.ParseFloat(string(npack[4:9]), 32)
	if err != nil {
		return
	}
	w.speed = w.speed / 100

	w.avg, err = strconv.ParseFloat(string(npack[9:14]), 32)
	if err != nil {
		return
	}
	w.avg = w.avg / 100

	return
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Error: no hostname")
	}
	server := os.Args[1]

	var port int
	var err error
	if len(os.Args) > 2 {
		port, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Error: port poorly formatted")
		}
	} else {
		port = 7756

	}

	ping := []byte("PING")
	poll := []byte{0x01}
	info := []byte{0x0D}

	conreq := make([]byte, 12)
	copy(conreq[0:6], []byte("CRQST"))
	conreq[9] = byte(port >> 8)
	conreq[10] = byte(port & 0xFF)
	conreq[11] = 0
	for i := 0; i < 11; i++ {
		conreq[11] ^= conreq[i]
	}

	// Open our port
	serveraddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:7754", server))
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
	_, err = Conn.Write(conreq)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "CRQST")
	Conn.Close()

	// Start acutual connection
	serveraddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		log.Fatal(err)
	}
	Conn, err = net.DialUDP("udp", localaddr, serveraddr)
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()

	//Keep Alive and polling
	go func() {
		i := 0
		for ; ; i++ {
			i = i % 5
			switch i {
			case 0:
				_, err = Conn.Write(info)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "INFO")
			case 1:
				_, err = Conn.Write(poll)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "POLL")
			default:
				_, err = Conn.Write(ping)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "PING")
			}
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	buf := make([]byte, 1024)
	stname := ""
	for {
		err = Conn.SetReadDeadline(time.Now().Add(time.Second * 20))
		n, addr, err := Conn.ReadFromUDP(buf)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			log.Println(err)
			continue
		}
		if err != nil {
			log.Fatal(err)
		}

		switch buf[0] {
		case 'P':
			log.Printf("%s:%d - %s\n", addr.IP.String(), addr.Port, string(buf[0:n]))
		case 0x01:
			log.Printf("%s:%d - WINDDATA\n", addr.IP.String(), addr.Port)
			w, err := parseWindData(buf)
			if err != nil {
				log.Println(err)
				continue
			}
			const TIMEFMT = "2006-01-02 15:04:05"
			fmt.Printf("%s %s %6.2f %03.0f %6.2f %03d %1d\n", stname, time.Now().Format(TIMEFMT), w.speed, w.head, w.avg, w.batt, w.stow)
		case 0x0D:
			log.Printf("%s:%d - STATIONDATA\n", addr.IP.String(), addr.Port)
			stname = string(buf[0 : n-1])
		}

	}

}
