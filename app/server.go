package main

import (
	"fmt"
	"net"
	"os"
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

    conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

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
    } else {
        conn.Write([]byte("HTTP/1.1 404 Not Found \r\n\r\n"))
    }
}
