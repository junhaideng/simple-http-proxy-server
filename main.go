package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	MAX_BUFF_SIZE = 1024
)

// NOW only HTTP
func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

// 处理链接
func handleConn(conn net.Conn) {
	defer conn.Close()
	var request = make([]byte, MAX_BUFF_SIZE)

	// 从conn中读取请求数据
	n, err := conn.Read(request)
	if err != nil {
		fmt.Println("read request error: ", err)
		return
	}

	reader := bytes.NewReader(request[:n])
	r := bufio.NewReader(reader)

	// 读取第一行请求数据
	s, err := r.ReadString('\n')
	if err != nil {
		fmt.Println("read string error: ", err)
		return
	}

	uri := strings.Split(s, " ")[1]

	// find  hostname, for example  httpbin.org
	// instead of http://httpbin.org/
	if strings.Index(uri, "http://") > -1 {
		uri = uri[7:]
	}

	// get server hostname
	pos := strings.Index(uri, "/")
	var hostname = uri
	if pos > -1 {
		hostname = uri[:pos]
	}
	// fmt.Println("hostname: ", hostname)

	// 获取到主机，以及端口号
	colon := strings.Index(hostname, ":")
	var host, port string
	if colon > -1 {
		host = hostname[:colon]
		port = hostname[colon+1:]
	} else {
		host = hostname
		port = "80"
	}
	fmt.Printf("host: %s, port: %s\n", host, port)

	// 建立到想要请求的服务端的连接
	c, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 30*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将需要请求的数据转发一份
	_, err = c.Write(request)
	if err != nil {
		fmt.Println("write request error: ", err)
		return
	}

	var buff [512]byte
	for {
		n, err := c.Read(buff[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		
		_, err = conn.Write(buff[:n])
		if err != nil {
			fmt.Println("write to client error: ", err)
			return
		}
	}
}
