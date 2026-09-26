package worker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Server struct {
	network  string
	addr     string
	executor *Executor
	listener net.Listener
	mu       sync.Mutex
	closed   bool
}

func NewServer(network, addr string, executor *Executor) *Server {
	return &Server{
		network:  network,
		addr:     addr,
		executor: executor,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("server already closed")
	}

	if s.network == "unix" {
		if dir := filepath.Dir(s.addr); dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		_ = os.Remove(s.addr)
	}

	l, err := net.Listen(s.network, s.addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen on %s:%s: %w", s.network, s.addr, err)
	}
	s.listener = l
	s.mu.Unlock()

	// Enforce 0660 permission on unix socket as required by GEMINI.md
	if s.network == "unix" && runtime.GOOS == "linux" {
		_ = os.Chmod(s.addr, 0660)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			s.mu.Lock()
			isClosed := s.closed
			s.mu.Unlock()
			if isClosed {
				return nil
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				// connection error
			}
			return
		}

		var cmd Command
		if err := json.Unmarshal(line, &cmd); err != nil {
			resp := Response{
				Success: false,
				Error:   fmt.Sprintf("malformed json request: %v", err),
			}
			_ = s.writeResponse(conn, resp)
			continue
		}

		data, err := s.executor.Execute(cmd)
		resp := Response{
			ID:      cmd.ID,
			Success: err == nil,
			Data:    data,
		}
		if err != nil {
			resp.Error = err.Error()
		}

		_ = s.writeResponse(conn, resp)
	}
}

func (s *Server) writeResponse(conn net.Conn, resp Response) error {
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	_, err = conn.Write(bytes)
	return err
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if s.listener != nil {
		err := s.listener.Close()
		if s.network == "unix" {
			_ = os.Remove(s.addr)
		}
		return err
	}
	return nil
}
