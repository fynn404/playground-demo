// Go Playground 后端服务
// 提供代码执行和格式化功能
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

// CodeRequest 定义了接收代码的请求结构
type CodeRequest struct {
	Code string `json:"code"` // 用户提交的代码
}

// CodeResponse 定义了代码执行的响应结构
type CodeResponse struct {
	Output string `json:"output"`          // 代码执行的输出结果
	Error  string `json:"error,omitempty"` // 执行过程中的错误信息（如果有）
}

// FormatResponse 定义了代码格式化的响应结构
type FormatResponse struct {
	FormattedCode string `json:"formattedCode"`   // 格式化后的代码
	Error         string `json:"error,omitempty"` // 格式化过程中的错误信息（如果有）
}

func main() {
	// 注册路由处理函数
	http.HandleFunc("/", serveStatic)            // 处理静态文件请求
	http.HandleFunc("/run", handleCodeExecution) // 处理代码执行请求
	http.HandleFunc("/format", handleCodeFormat) // 处理代码格式化请求

	// 启动服务器
	port := ":8081"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// serveStatic 处理静态文件的请求
// 如果请求根路径，返回 index.html
// 其他静态文件请求直接从 static 目录提供服务
func serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}
	http.FileServer(http.Dir("static")).ServeHTTP(w, r)
}

// handleCodeExecution 处理代码执行请求
// 1. 验证请求方法
// 2. 解析请求体中的代码
// 3. 执行代码并返回结果
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

// handleCodeFormat 处理代码格式化请求
// 1. 验证请求方法
// 2. 解析请求体中的代码
// 3. 使用 gofmt 格式化代码并返回结果
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

// formatCode 使用 gofmt 格式化 Go 代码
// 1. 创建临时目录和文件
// 2. 将代码写入临时文件
// 3. 使用 gofmt 格式化代码
// 4. 返回格式化后的代码
func formatCode(code string) (string, error) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "goformat")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建临时文件并写入代码
	tmpFile := filepath.Join(tmpDir, "code.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write code file: %v", err)
	}

	// 运行 gofmt 格式化代码
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

// executeCode 执行 Go 代码并返回执行结果
// 1. 创建临时目录和文件
// 2. 将代码写入临时文件
// 3. 执行代码并捕获输出
// 4. 返回执行结果
func executeCode(code string) (string, error) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "goplayground")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建临时文件并写入代码
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write code file: %v", err)
	}

	// 执行代码
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

	// 收集程序输出
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
