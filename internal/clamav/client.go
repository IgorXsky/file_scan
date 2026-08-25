package clamav

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type Result struct {
	Clean     bool   `json:"clean"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Client struct {
	addr    string
	timeout time.Duration
}

func NewClient(addr string, timeout time.Duration) *Client {
	return &Client{addr: addr, timeout: timeout}
}

func (c *Client) Ping() error {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return fmt.Errorf("clamav connect: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("clamav write: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		resp := strings.TrimRight(scanner.Text(), "\x00")
		if resp == "PONG" {
			return nil
		}
		return fmt.Errorf("clamav unexpected response: %s", resp)
	}

	return fmt.Errorf("clamav no response")
}

func (c *Client) ScanStream(reader io.Reader) (Result, error) {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return Result{}, fmt.Errorf("clamav connect: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return Result{}, fmt.Errorf("clamav write command: %w", err)
	}

	buf := make([]byte, 2048)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			size := make([]byte, 4)
			binary.BigEndian.PutUint32(size, uint32(n))
			if _, err := conn.Write(size); err != nil {
				return Result{}, fmt.Errorf("clamav write chunk size: %w", err)
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return Result{}, fmt.Errorf("clamav write chunk: %w", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return Result{}, fmt.Errorf("read input: %w", readErr)
		}
	}

	// Send zero-length chunk to signal end
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return Result{}, fmt.Errorf("clamav write terminator: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return Result{}, fmt.Errorf("clamav no response")
	}

	return parseResponse(scanner.Text()), nil
}

func parseResponse(resp string) Result {
	resp = strings.TrimRight(resp, "\x00")

	if strings.HasSuffix(resp, "OK") {
		return Result{Clean: true}
	}

	if strings.Contains(resp, "FOUND") {
		parts := strings.SplitN(resp, ":", 2)
		sig := ""
		if len(parts) > 1 {
			sig = strings.TrimSpace(parts[1])
			sig = strings.TrimSuffix(sig, "FOUND")
			sig = strings.TrimSpace(sig)
		}
		return Result{Clean: false, Signature: sig}
	}

	return Result{Error: resp}
}
