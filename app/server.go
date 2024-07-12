package main

import (
	"fmt"
	"net"
	"os"
)

func parseRequest(request string)(string,string){
	var method, path string
	fmt.Sscanf(request, "%s %s HTTP/1.1", &method, &path)
	return method, path
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	req := make([]byte, 1024)
	conn.Read(req)
	method, path := parseRequest(string(req))
	fmt.Println("Method:", method)
	fmt.Println("Path:", path)

	if method == "GET" && path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
