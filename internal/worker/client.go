package worker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Client struct {
	network string
	addr    string
	conn    net.Conn
	reader  *bufio.Reader
	mu      sync.Mutex
}

func NewClient(network, addr string) *Client {
	return &Client{
		network: network,
		addr:    addr,
	}
}

func (c *Client) connect() error {
	if c.conn != nil {
		return nil
	}

	conn, err := net.DialTimeout(c.network, c.addr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to root worker at %s:%s: %w", c.network, c.addr, err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func (c *Client) Call(action string, params map[string]string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.connect(); err != nil {
		return "", err
	}

	cmd := Command{
		ID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		Action: action,
		Params: params,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return "", err
	}
	payload = append(payload, '\n')

	_ = c.conn.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.conn.Write(payload); err != nil {
		c.conn.Close()
		c.conn = nil
		return "", fmt.Errorf("failed to send command to worker: %w", err)
	}

	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return "", fmt.Errorf("failed to read response from worker: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return "", fmt.Errorf("failed to decode worker response: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("worker error: %s", resp.Error)
	}

	return resp.Data, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}
