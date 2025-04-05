package main

import (
	"fmt"
	"net"
	"os"
    "io"
    "bytes"
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

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting connection: ", err.Error())
            os.Exit(1)
        }
        defer conn.Close()

        go handleRequest(conn)
    }
}

func handleRequest(conn net.Conn) {
    req := make([]byte, 100)
    conn.Read(req)

    p_req := make([][][]byte, 0)
    vals := make([][]byte, 0)
    beg_idx := 0
    for i:=0; i < len(req); i++ {
        if req[i] == ' ' {
            vals = append(vals, req[beg_idx:i])
            beg_idx = i+1
        } else if req[i] == '\r' && len(req) >= i+1 && req[i+1] == '\n' {
            vals = append(vals, req[beg_idx:i])
            p_req = append(p_req, vals)
            vals = make([][]byte, 0)
            i++ // catch up to line feed
            beg_idx = i+1
        }
    }

    if bytes.Equal(p_req[0][1], []byte("/")) {
        conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
    } else if bytes.Equal(p_req[0][1][0:6], []byte("/echo/")) {
        content := p_req[0][1][6:]
        out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
        conn.Write([]byte(out))
    } else if bytes.Equal(p_req[0][1], []byte("/user-agent")) {
        for i:=0; i < len(p_req); i++ {
            if bytes.Equal(p_req[i][0], []byte("User-Agent:")) {
                agent := p_req[i][1]
                out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)
                conn.Write([]byte(out))
                break
            }
        }
    } else if bytes.Equal(p_req[0][1][0:7], []byte("/files/")) {
        filename := string(p_req[0][1][7:])
        f, err := os.Open("/tmp/"+filename)
        if err != nil {
            conn.Write([]byte("HTTP/1.1 400 Not Found\r\n\r\n"))
        } else {
            content := make([]byte, 0)
            _, err := f.Read(content)
            for err != io.EOF {
                c := make([]byte, 10)
                _, err = f.Read(c)
                content = append(content, c...)
            }
            
            out := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
            conn.Write([]byte(out))
        }
    } else {
        conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
    }
}
