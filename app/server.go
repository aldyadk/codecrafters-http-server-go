package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var dir string

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

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func handleConnection(conn net.Conn) {
	req := make([]byte, 1024)
	conn.Read(req)
	request := string(req)
	method, path, headers, body := parseRequest(string(request))
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
			body = headers["User-Agent"]
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)))
		} else {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	} else if method == "GET" && pathA == "echo" && pathB != "" {
		body := []byte(body)
		compressedBody, err := gzipCompress(body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(compressedBody)

		if headers["Accept-Encoding"] != "" && strings.Contains(headers["Accept-Encoding"], "gzip") {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: %s\r\nContent-Length: %d\r\n\r\n%s", "gzip", len(compressedBody), compressedBody)))
		} else {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathB), pathB)))
		}
	} else if method == "POST" && pathA == "files" && pathB != "" {
		filePath := dir + string(os.PathSeparator) + pathB
		if _, err := os.Stat(filePath); err == nil {
			fmt.Println("File already exists:", filePath)
			conn.Write([]byte("HTTP/1.1 405 Not Allowed\r\n\r\n"))
			return
		}
		fmt.Println(body)
		content := []byte(strings.Trim(body, "\x00"))
		fmt.Println(content)
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	} else if method == "GET" && pathA == "files" && pathB != "" {
		filePath := dir + string(os.PathSeparator) + pathB
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			fmt.Println("Error getting file info:", err)
			return
		}
		data := make([]byte, stat.Size())
		_, err = file.Read(data)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", stat.Size(), string(data))))
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
	directory := flag.String("directory", ".", "Directory to process")
	flag.Parse()
	if *directory != "" {
		dir = *directory
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
