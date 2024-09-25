package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	return aof, nil
}

// Write method for appending commands to AOF
func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// Convert the Value to a string format (RESP or plain text)
	data := convertValueToString(value)

	// Write the data to the file
	_, err := aof.file.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write to AOF: %v", err)
	}

	// Ensure the data is synced to disk
	err = aof.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync AOF: %v", err)
	}

	return nil
}

// Close method for Aof
func (aof *Aof) Close() {
	aof.file.Close()
}

// Helper function to convert Value into a string format
func convertValueToString(value Value) string {
	if value.typ == "array" {
		result := "*"
		result += fmt.Sprintf("%d\r\n", len(value.array))
		for _, v := range value.array {
			result += fmt.Sprintf("$%d\r\n%s\r\n", len(v.bulk), v.bulk)
		}
		return result
	}
	return ""
}

// New Read method for replaying AOF commands
func (aof *Aof) Read(callback func(Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// Reset the reader to start from the beginning of the file
	aof.file.Seek(0, 0)
	respReader := NewResp(aof.file) // Use the Resp struct for parsing RESP

	for {
		// Read and parse the next value from the AOF file
		value, err := respReader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return fmt.Errorf("failed to read from AOF: %v", err)
		}

		// Call the callback with the parsed value
		callback(value)
	}

	return nil
}
