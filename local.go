package main

import (
	"log"
	"net"
	"proxy"
)

//
//
//     bs ----> client ====> proxy ----> target
//     bs <---- client <==== proxy <---- target


func main()  {
	log.SetFlags(log.LstdFlags|log.Lshortfile)
	log.Print("proxy start listen 0.0.0.0:8081")
	l, err := net.Listen("tcp", "0.0.0.0:8081")
	if err != nil {
		return
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go forward(client)
	}

}

func forward(client net.Conn)  {
	var b [1024]byte
	n, err := client.Read(b[:])
	client.Write([]byte{0x05, 0x00})

	n, err = client.Read(b[:])
	client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	if err != nil {
		return
	}

	server1 := proxy.Server{"35.185.165.210", "8082", "Client"}
	//server1 := proxy.Server{"127.0.0.1", "8082", "Client"}
	host, port := server1.AnaHost(b, n)
	log.Print("Dial start ", host, ":", port)

	server := server1.Dial()
	server.Write(b[:n])


	go server1.Shovel(client, server, server.RemoteAddr())
}