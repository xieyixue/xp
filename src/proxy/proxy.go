package proxy

import (
	"net"
	"time"
	"log"
	"strconv"
	"errors"
	"crypto/aes"
	"crypto/cipher"
)

var ErrShortWrite = errors.New("short write")
var EOF = errors.New("EOF")

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Mode string `json:"mode"`
}

func (s Server) Dial() net.Conn {
	server, err := net.DialTimeout("tcp", net.JoinHostPort(s.Host, s.Port), 3 * time.Second )
	if err != nil {
		log.Panic(err)
	}
	return server
}

func (s Server) Listen() (error){
	l, err := net.Listen("tcp", net.JoinHostPort(s.Host, s.Port))
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go s.HandleClientSocks(client)
	}

}

func (s Server) AnaHost(b [1024]byte, n int) (host string, port string) {
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

func (s Server) Shovel(src, dst net.Conn, n net.Addr) error{
	defer src.Close()
	defer dst.Close()
	errch := make(chan error, 1)

	go s.chanCopy(errch, src, dst, n)
	go s.chanCopy(errch, dst, src, n)

	for i := 0; i < 2; i++ {
		if err := <-errch; err != nil {
			// If this returns early the second func will push into the
			// buffer, and the GC will clean up
			return err
		}
	}
	return nil

}
func (s Server) chanCopy(e chan error, dst, src net.Conn,  n net.Addr) {
	_, err := s.Copy(dst, src, n)
	e <- err
}


func (s Server) Copy(src , dst net.Conn, n net.Addr) (written int64, err error){
	var v int = 2
	if s.Mode == "Client" {
		if n == src.RemoteAddr() {
			v = 0
			//log.Println("client read proxy")
		}
		
		if n == dst.RemoteAddr() {
			v = 1
			//log.Println("client write proxy")
		}
	}
	if s.Mode == "Server" {
		if n == dst.RemoteAddr() {
			v = 1
			//log.Println("proxy write client")
		}
		if n == src.RemoteAddr() {
			v = 0
			log.Println("proxy read client", v)
		}
	}

	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			//buf = buf[:nr]
			if v != 2 {
				enBuf := s.encode(buf[:nr], v)
				buf = enBuf
			}else {
				buf = buf[:nr]
			}
			nw, ew := dst.Write(buf)
			if nw >0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	//log.Println(s.Mode, ":", written, "=====", err, "OK")
	return
}

func (s Server) HandleClientSocks(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	var b [1024]byte

	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	if b[0] == 0x05 {
		log.Println(b[:n])
		host, port := s.AnaHost(b, n)
		server, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
		log.Print("dial start ", host, ":", port)
		if err != nil {
			log.Println(err)
			return
		}
		s.Shovel(client, server, client.RemoteAddr())
	}
}

func (s Server) encode(buf []byte, v int)  (enBuf []byte){
	key_text := "astaxie12798akljzmknm.ahkjkljl;k"
	var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	c, err := aes.NewCipher([]byte(key_text))
	if err != nil {
		log.Println(err)
	}


	if v == 1 {
		cfb := cipher.NewCFBEncrypter(c, commonIV)
		ciphertext := make([]byte, len(buf))
		cfb.XORKeyStream(ciphertext, buf)
		return ciphertext
	}else {
		cfbdec := cipher.NewCFBDecrypter(c, commonIV)
		plaintextCopy := make([]byte, len(buf))
		cfbdec.XORKeyStream(plaintextCopy, buf)
		return plaintextCopy
	}
	
}