package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func parseRequest(request string) (string, string, map[string]string, string) {
	var method, path, body string
	headers := make(map[string]string)
	lines := strings.Split(request, "\r\n")
	fmt.Sscanf(lines[0], "%s %s HTTP/1.1", &method, &path)

	var i int
	for i = 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		headers[headerParts[0]] = headerParts[1]
	}

	body = strings.Join(lines[i+1:], "\r\n")

	return method, path, headers, body
}

func parsePath(path string) []string {
	return strings.Split(path, "/")

}

func handleConnection(conn net.Conn) {
	req := make([]byte, 1024)
	conn.Read(req)
	request := string(req)
	method, path, headers, _ := parseRequest(string(request))
	paths := parsePath(path)
	var pathA, pathB string
	pathA = paths[1]
	if len(paths) > 2 {
		pathB = paths[2]
	}

	if method == "GET" && path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if method == "GET" && pathA == "user-agent" {
		if headers["User-Agent"] != "" {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(headers["User-Agent"]), headers["User-Agent"])))
		} else {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	} else if method == "GET" && pathA == "echo" && pathB != "" {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathB), pathB)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
