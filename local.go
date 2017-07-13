package main

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var proxys = map[int]string{}
var CounterMap = map[string]int{}
var countChan = make(chan string, 1000)

func shutdown(c chan os.Signal, server net.Listener, m map[string]int)  {
	s := <-c
	fmt.Println("Got signal:", s)
	server.Close()
	for k, v := range m{
		log.Print(k, "-->", v)
	}
	os.Exit(0)
}

func Counter()  {
	for  {
		s_host := <- countChan
		if CounterMap[s_host] == 0 {
			CounterMap[s_host] = 1

		}else {
			CounterMap[s_host] += 1
		}
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	proxys[0] = "default"
	proxys[1] = "127.0.0.1:5555"
	proxys[2] = "127.0.0.1:5557"
	proxys[2] = "35.185.165.210:5555"

	log.SetFlags(log.LstdFlags|log.Lshortfile)
	log.Print("proxy start listen 0.0.0.0:8081")
	l, err := net.Listen("tcp", ":8081")
	m := map[string]int{}
	go shutdown(c, l, m)
	go Counter()

	go Config(m)
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go forward(m, client)

	}
}

func forward(m map[string]int, client net.Conn) {

	defer client.Close()
	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err, n)
		return
	}

	client.Write([]byte{0x05, 0x00})
	n, err = client.Read(b[:])
	if err != nil {
		log.Println(err, n)
		return
	}
	host, port := AnaHost(b, n)

	d_host, d_port := GetServer(m, host, port)
	server, err := net.DialTimeout("tcp", net.JoinHostPort(d_host, d_port ), 6*time.Second)
	log.Print("connect to ", d_host, ":", d_port)
	if err != nil {
		log.Println(err)
		m[d_host] = 1
		return
	}
	defer server.Close()

	if d_host == host {
		client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}else {
		server.Write([]byte{0x05, 0x01, 0x00})
		s := b[:n]
		var b [1024]byte
		n, err = server.Read(b[:])
		server.Write(s)

	}
	Shovel(server, client)

}

// proxy between two sockets
func Shovel(local, remote io.ReadWriteCloser) error {
	errch := make(chan error, 1)

	go chanCopy(errch, local, remote)
	go chanCopy(errch, remote, local)

	for i := 0; i < 2; i++ {
		if err := <-errch; err != nil {
			// If this returns early the second func will push into the
			// buffer, and the GC will clean up
			return err

		}
	}
	return nil
}

// copy between pipes, sending errors to channel
func chanCopy(e chan error, dst, src io.ReadWriter) {
	_, err := io.Copy(dst, src)
	e <- err
}

func AnaHost(b [1024]byte, n int) (host string, port string) {
	switch b[3] {
	case 0x01:
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03:
		host = string(b[5 : n-2])
	case 0x04:
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	case 0x05:
		host = string(b[8 : n-2])
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	return host, port
}


func GetServer(m map[string]int, s_host, s_port string) (d_host, d_port string) {
	countChan <- s_host
	target := m[s_host]
	if target != 0 {
		proxy := proxys[target]
		data := strings.Split(proxy, ":")
		if proxy != "" {
			log.Print("代理 ", data[0], ":", data[1])
			return data[0], data[1]
		}
	}
	m[s_host] = 0
	log.Print("直连 ", s_host, ":", s_port)
	return s_host, s_port

}

func Config(m map[string]int)  {

	log.Print("config server start listen 0.0.0.0:8082")
	l, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		var b [1024]byte
		for {
			n, err := client.Read(b[:])
			if err != nil {
				break
			}
			data := strings.Fields(string(b[:n]))
			log.Print(string(b[:n]))
			if len(data) == 0 {
				break
			}
			if data[0] == "set" {
				target, err := strconv.Atoi(data[2])
				m[data[1]] = target
				if err != nil {
					log.Print(err, " ", data[2])
				}
			}
			if data[0] == "get" {
				for k, v := range m{
					sl := fmt.Sprintf("%s==>%d-->%d\n", k, v, CounterMap[k])
					client.Write([]byte(sl))
				}
			}
			if data[0] == "close" {
				break
			}
		}
		log.Print("closed")
		client.Close()

	}
}