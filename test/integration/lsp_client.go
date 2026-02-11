// Package integration provides an LSP protocol client for integration testing.
package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Client speaks the LSP protocol over stdio to a tekton-lsp binary.
type Client struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	nextID   atomic.Int64
	t        *testing.T
	mu       sync.Mutex
	notifs   []json.RawMessage
	initDone bool
}

// StartServer launches the tekton-lsp binary and returns a client.
func StartServer(t *testing.T, binary string) *Client {
	t.Helper()
	cmd := exec.Command(binary)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	cmd.Stderr = nil // discard

	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	c := &Client{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReaderSize(stdout, 1<<20),
		t:      t,
	}

	t.Cleanup(func() {
		_ = c.Shutdown()
		_ = cmd.Wait()
	})

	return c
}

// Initialize sends initialize + initialized.
func (c *Client) Initialize() map[string]any {
	resp := c.Request("initialize", map[string]any{
		"capabilities": map[string]any{},
		"processId":    nil,
		"rootUri":      "file:///tmp/tekton-test",
	})
	c.Notify("initialized", map[string]any{})
	c.initDone = true
	return resp
}

// OpenFile sends textDocument/didOpen.
func (c *Client) OpenFile(uri, content string) {
	c.Notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": "yaml",
			"version":    1,
			"text":       content,
		},
	})
	// Small sleep to let server process
	time.Sleep(50 * time.Millisecond)
}

// Completion sends textDocument/completion.
func (c *Client) Completion(uri string, line, char uint32) map[string]any {
	return c.Request("textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": line, "character": char},
	})
}

// Hover sends textDocument/hover.
func (c *Client) Hover(uri string, line, char uint32) map[string]any {
	return c.Request("textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": line, "character": char},
	})
}

// Definition sends textDocument/definition.
func (c *Client) Definition(uri string, line, char uint32) map[string]any {
	return c.Request("textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": line, "character": char},
	})
}

// DocumentSymbols sends textDocument/documentSymbol.
func (c *Client) DocumentSymbols(uri string) map[string]any {
	return c.Request("textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
}

// Formatting sends textDocument/formatting.
func (c *Client) Formatting(uri string, tabSize int) map[string]any {
	return c.Request("textDocument/formatting", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"options":      map[string]any{"tabSize": tabSize, "insertSpaces": true},
	})
}

// CodeAction sends textDocument/codeAction.
func (c *Client) CodeAction(uri string, startLine, startChar, endLine, endChar uint32) map[string]any {
	return c.Request("textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"range": map[string]any{
			"start": map[string]any{"line": startLine, "character": startChar},
			"end":   map[string]any{"line": endLine, "character": endChar},
		},
		"context": map[string]any{
			"diagnostics": []any{},
		},
	})
}

// Shutdown sends shutdown + exit.
func (c *Client) Shutdown() error {
	c.Request("shutdown", nil)
	c.Notify("exit", nil)
	c.stdin.Close()
	return nil
}

// Request sends a request and waits for the response.
func (c *Client) Request(method string, params any) map[string]any {
	id := c.nextID.Add(1)
	c.send(method, params, &id)
	return c.readResponse(id)
}

// Notify sends a notification (no response expected).
func (c *Client) Notify(method string, params any) {
	c.send(method, params, nil)
}

func (c *Client) send(method string, params any, id *int64) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	if id != nil {
		msg["id"] = *id
	}

	body, err := json.Marshal(msg)
	if err != nil {
		c.t.Fatalf("marshal: %v", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		c.t.Fatalf("write header: %v", err)
	}
	if _, err := c.stdin.Write(body); err != nil {
		c.t.Fatalf("write body: %v", err)
	}
}

func (c *Client) readResponse(id int64) map[string]any {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg, err := c.readMessage()
		if err != nil {
			c.t.Fatalf("read message: %v", err)
		}

		// Check if this is the response we're waiting for.
		if msgID, ok := msg["id"]; ok {
			var respID int64
			switch v := msgID.(type) {
			case float64:
				respID = int64(v)
			case json.Number:
				respID, _ = v.Int64()
			}
			if respID == id {
				return msg
			}
		}

		// Otherwise it's a notification — store it.
		raw, _ := json.Marshal(msg)
		c.notifs = append(c.notifs, raw)
	}
	c.t.Fatalf("timeout waiting for response id=%d", id)
	return nil
}

func (c *Client) readMessage() (map[string]any, error) {
	// Read headers until blank line.
	var contentLength int
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return msg, nil
}
