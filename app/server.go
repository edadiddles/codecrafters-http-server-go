package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	dir := ""
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		defer conn.Close()

		go handleRequest(conn, dir)
	}
}

func handleRequest(conn net.Conn, dir string) {
	req := make([]byte, 800)
	conn.Read(req)
	p_req := make([][][]byte, 0)
	vals := make([][]byte, 0)
	beg_idx := 0
	is_content_length := false
	content_length := 0
	for i := 0; i < len(req); i++ {
		if req[i] == '\x00' {
			vals = append(vals, req[beg_idx:i])
			p_req = append(p_req, vals)
			break
		} else if req[i] == ' ' {
			if bytes.Equal(req[beg_idx:i], []byte("Content-Length:")) {
				is_content_length = true
			}
			vals = append(vals, req[beg_idx:i])
			beg_idx = i + 1
		} else if req[i] == '\r' && len(req) >= i+1 && req[i+1] == '\n' {
			if is_content_length {
				is_content_length = false
				content_length, _ = strconv.Atoi(string(req[beg_idx:i]))
			}
			vals = append(vals, req[beg_idx:i])
			p_req = append(p_req, vals)
			vals = make([][]byte, 0)
			i++ // catch up to line feed
			beg_idx = i + 1
			if req[i+1] == '\r' && len(req) >= i+2 && req[i+2] == '\n' {
				i++
				i++
				beg_idx = i + 1
				vals = append(vals, req[beg_idx:beg_idx+content_length])
				p_req = append(p_req, vals)
				beg_idx += content_length
				break
			}
		}
	}

	if bytes.Equal(p_req[0][1], []byte("/")) {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if bytes.Equal(p_req[0][1][0:6], []byte("/echo/")) {
		content := p_req[0][1][6:]
		out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
		conn.Write([]byte(out))
	} else if bytes.Equal(p_req[0][1], []byte("/user-agent")) {
		for i := 0; i < len(p_req); i++ {
			if bytes.Equal(p_req[i][0], []byte("User-Agent:")) {
				agent := p_req[i][1]
				out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)
				conn.Write([]byte(out))
				break
			}
		}
	} else if bytes.Equal(p_req[0][1][0:7], []byte("/files/")) {
		filename := string(p_req[0][1][7:])
		if bytes.Equal(p_req[0][0], []byte("GET")) {
			f, err := os.Open(dir + filename)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				content := make([]byte, 0)
				_, err := f.Read(content)
				for err != io.EOF {
					c := make([]byte, 1)
					_, err = f.Read(c)
					if !bytes.Equal(c, []byte(string('\x00'))) {
						content = append(content, c...)
					}
				}

				out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
				conn.Write([]byte(out))
			}
		} else if bytes.Equal(p_req[0][0], []byte("POST")) {
			content := p_req[len(p_req)-1][0]
			os.WriteFile(dir+filename, content, 0666)
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
