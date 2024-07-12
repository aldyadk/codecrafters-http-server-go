package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func parseRequest(request string) (string, string) {
	var method, path string
	fmt.Sscanf(request, "%s %s HTTP/1.1", &method, &path)
	return method, path
}

func parsePath(path string) []string {
	return strings.Split(path, "/")

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
	paths := parsePath(path)
	var pathA, pathB string
	pathA = paths[1]
	if len(paths) > 2 {
		pathB = paths[2]
	}

	if method == "GET" && path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if method == "GET" && pathA == "echo" && pathB != "" {
		fmt.Println(pathB)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathB), pathB)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
