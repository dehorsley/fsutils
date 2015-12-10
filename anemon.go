package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type windData struct {
	time             time.Time
	name             string
	stow, batt       int
	head, speed, avg float64
}

func parseWindData(pack []byte) (w windData, err error) {
	if pack[0] == 1 {
		err = errors.New("Not valid packet")
	}

	p := make([]byte, len(pack))
	copy(p, pack)
	for i := 1; i < 14; i++ {
		p[i] += '0'
	}

	w.head, err = strconv.ParseFloat(string(p[1:4]), 32)
	if err != nil {
		return
	}

	w.speed, err = strconv.ParseFloat(string(p[4:9]), 32)
	if err != nil {
		return
	}
	w.speed = w.speed / 100

	w.avg, err = strconv.ParseFloat(string(p[9:14]), 32)
	if err != nil {
		return
	}
	w.avg = w.avg / 100

	w.batt = int(p[14])
	w.stow = int(p[15])

	w.time = time.Date(2000+int(p[16]), time.Month(p[17]), int(p[18]), int(p[19]), int(p[20]), int(p[21]), 0, time.UTC)

	w.name = string(p[23 : 23+int(p[22])])

	return
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Error: no hostname")
	}
	server := os.Args[1]
	port := 7755

	id := rand.Int31()
	idh := byte((id & 0xFF00) >> 8)
	idl := byte(id & 0x00FF)

	ping := []byte{idh, idl, 'P', 'I', 'N', 'G'}
	poll := []byte{idh, idl, 0x01}

	conreq := make([]byte, 14)
	conreq[0] = idh
	conreq[1] = idl
	copy(conreq[2:7], []byte("CRQST"))
	conreq[13] = 0
	for i := 0; i < 11; i++ {
		conreq[13] ^= conreq[i]
	}

	serveraddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		log.Fatal(err)
	}
	localaddr, err := net.ResolveUDPAddr("udp", ":7756")
	if err != nil {
		log.Fatal(err)
	}

	Conn, err := net.DialUDP("udp", localaddr, serveraddr)
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()

	_, err = Conn.Write(conreq)
	log.Printf("%s:%d - %s\n", "127.0.0.1", localaddr.Port, "CRQST")
	if err != nil {
		log.Fatal(err)
	}

	//Keep Alive and polling
	go func() {
		i := 0
		for ; ; i++ {
			i = i % 5
			switch i {
			case 0:
				_, err = Conn.Write(poll)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "POLL")
			default:
				_, err = Conn.Write(ping)
				log.Printf("127.0.0.1:%d - %s\n", localaddr.Port, "PING")
			}
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	buf := make([]byte, 1024)
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
			fmt.Printf("%s %s %6.2f %03.0f %6.2f %03d %1d\n", w.name, time.Now().Format(TIMEFMT), w.speed, w.head, w.avg, w.batt, w.stow)
			// Could use this, but pcfs time is more likely to be correct
			// fmt.Printf("%s %s %6.2f %03.0f %6.2f %03d %1d\n", w.name, w.time.Format(TIMEFMT), w.speed, w.head, w.avg, w.batt, w.stow)
		}

	}

}
