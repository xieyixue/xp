package main

import "fmt"
import "net"
import "time"
import "sync"
import "strconv"
import "net/http"
import "io/ioutil"
import "golang.org/x/net/proxy"

var BestAddr = ""

func Socks5Client(addr string) (client *http.Client, err error) {

	dialer, err := proxy.SOCKS5("tcp", addr,
		nil,
		&net.Dialer {
			Timeout: 30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	)
	if err != nil {
		return
	}
	
	transport := &http.Transport{
		Proxy: nil,
		Dial: dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client = &http.Client {Transport: transport}
	return
}

func check(addr string) (err error){
	start := time.Now()
	client, err := Socks5Client(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	rsp, err := client.Get("http://myip.ipip.net")
	if err != nil {
		fmt.Println(err)
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
    rsp.Body.Close()
    if err != nil {
        fmt.Println(err)
    } else {
		end := time.Now()
		delta := end.Sub(start)
		if BestAddr == "" {
			BestAddr = addr
		}
		fmt.Printf("%s %s %s\n", addr, delta, string(b))
    }
	return
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
	}
	for a := 5555; a < 5575; a++ {
		addr := "127.0.0.1:" + strconv.Itoa(a)
		go func (addr string) {
			err := check(addr)
			if err == nil {
				wg.Done()
			}
		}(addr)
	 }
	 wg.Wait()
	 fmt.Println(BestAddr)
}