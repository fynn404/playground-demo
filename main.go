package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type CodeRequest struct {
	Code string `json:"code"`
}

type CodeResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type FormatResponse struct {
	FormattedCode string `json:"formattedCode"`
	Error         string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/", serveStatic)
	http.HandleFunc("/run", handleCodeExecution)
	http.HandleFunc("/format", handleCodeFormat)

	port := ":8081"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}
	http.FileServer(http.Dir("static")).ServeHTTP(w, r)
}

func handleCodeExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := executeCode(req.Code)
	resp := CodeResponse{Output: output}
	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCodeFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formattedCode, err := formatCode(req.Code)
	resp := FormatResponse{FormattedCode: formattedCode}
	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func formatCode(code string) (string, error) {
	// 创建一个临时文件来存储代码
	tmpDir, err := os.MkdirTemp("", "goformat")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "code.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write code file: %v", err)
	}

	// 运行 gofmt
	cmd := exec.Command("gofmt", tmpFile)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("format error: %v\n%s", err, stderr.String())
	}

	return out.String(), nil
}

func executeCode(code string) (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "goplayground")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file for the code
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write code file: %v", err)
	}

	// Run the code
	cmd := exec.Command("go", "run", tmpFile)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %v", err)
	}

	// Combine stdout and stderr
	output := make(chan string)
	go func() {
		combined := ""
		stdoutBytes, _ := io.ReadAll(stdout)
		stderrBytes, _ := io.ReadAll(stderr)
		combined += string(stdoutBytes)
		if len(stderrBytes) > 0 {
			combined += "\nError:\n" + string(stderrBytes)
		}
		output <- combined
	}()

	err = cmd.Wait()
	return <-output, err
}
