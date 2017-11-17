package main

import "io"
import "os"
import "net"
import "fmt"
import "time"
import "sync"


func PortForward (client net.Conn) {
	server, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "5574"), 3 * time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var wg sync.WaitGroup
    go func(client net.Conn, server net.Conn) {
        wg.Add(1)
        defer wg.Done()
        io.Copy(client, server)
        client.Close()
    }(client, server)
    go func(client net.Conn, server net.Conn) {
        wg.Add(1)
        defer wg.Done()
        io.Copy(server, client)
        server.Close()
    }(client, server)
	wg.Wait()
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6002")
	if err != nil {
		return
	}
	for {
		client, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go PortForward(client)
	}
}
