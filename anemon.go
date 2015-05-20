package main

import (
	"errors"
	"fmt"
	"log"
	"net"
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

type stationData struct {
	head, speed, avg float64
}

func main() {

	ping := []byte("PING")
	poll := []byte{0x01}
	info := []byte{0x0D}

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

	//Keep Alive
	go func() {
		i := 0
		for {
			i++
			i = i % 10
			switch i {
			case 0:
				_, err = Conn.Write(poll)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "POLL")
			case 1:
				_, err = Conn.Write(info)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "INFO")
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
	for {
		n, addr, err := Conn.ReadFromUDP(buf)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
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
			fmt.Println(w)
		case 0x0D:
			log.Printf("%s:%d - STATIONDATA\n", addr.IP.String(), addr.Port)
		}

		// err = Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	}

}
