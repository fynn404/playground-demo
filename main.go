// Go Playground 后端服务
// 提供代码执行和格式化功能
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 安全配置常量
const (
	// 代码执行超时时间（秒）
	execTimeout = 5
	// 最大代码长度（字节）
	maxCodeLength = 10 * 1024 // 10KB
	// 最大输出长度（字节）
	maxOutputLength = 1024 * 1024 // 1MB
	// Docker 容器名前缀
	containerPrefix = "playground-"
	// Docker 镜像名
	dockerImage = "playground-runner:latest"
)

// 禁止导入的包列表
var forbiddenPackages = []string{
	"os/exec", // 禁止执行系统命令
	"syscall", // 禁止系统调用
	"unsafe",  // 禁止不安全操作
	"net",     // 禁止网络操作
	"os",      // 禁止文件系统操作
	"plugin",  // 禁止加载插件
	"debug",   // 禁止调试功能
}

// 速率限制器
type RateLimiter struct {
	requests map[string]*TokenBucket
	mu       sync.Mutex
}

// 令牌桶
type TokenBucket struct {
	tokens   float64
	lastTime time.Time
	rate     float64
	capacity float64
}

// 创建新的速率限制器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*TokenBucket),
	}
}

// 检查请求是否允许
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.requests[ip]
	if !exists {
		bucket = &TokenBucket{
			tokens:   3,          // 初始令牌数
			lastTime: time.Now(), // 当前时间
			rate:     0.5,        // 每秒恢复的令牌数
			capacity: 3,          // 最大令牌数
		}
		rl.requests[ip] = bucket
	}

	// 计算从上次请求到现在应该恢复的令牌数
	now := time.Now()
	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens = bucket.tokens + elapsed*bucket.rate
	if bucket.tokens > bucket.capacity {
		bucket.tokens = bucket.capacity
	}

	// 如果没有足够的令牌，拒绝请求
	if bucket.tokens < 1 {
		return false
	}

	// 消耗一个令牌
	bucket.tokens--
	bucket.lastTime = now
	return true
}

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

var rateLimiter = NewRateLimiter()

func main() {
	// 确保 Docker 镜像已经构建
	if err := buildDockerImage(); err != nil {
		log.Fatalf("Failed to build Docker image: %v", err)
	}

	// 注册路由处理函数
	http.HandleFunc("/", serveStatic)            // 处理静态文件请求
	http.HandleFunc("/run", handleCodeExecution) // 处理代码执行请求
	http.HandleFunc("/format", handleCodeFormat) // 处理代码格式化请求

	// 启动服务器
	port := ":8081"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// buildDockerImage 构建 Docker 镜像
func buildDockerImage() error {
	cmd := exec.Command("docker", "build", "-t", dockerImage, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// 检查代码是否包含禁止的包导入
func checkForbiddenImports(code string) error {
	for _, pkg := range forbiddenPackages {
		if strings.Contains(code, fmt.Sprintf("import %q", pkg)) ||
			strings.Contains(code, fmt.Sprintf(`import "%s"`, pkg)) ||
			strings.Contains(code, fmt.Sprintf("import (%q)", pkg)) {
			return fmt.Errorf("forbidden package import: %s", pkg)
		}
	}
	return nil
}

// 验证代码安全性
func validateCode(code string) error {
	// 检查代码长度
	if len(code) > maxCodeLength {
		return fmt.Errorf("code length exceeds maximum allowed (%d bytes)", maxCodeLength)
	}

	// 检查禁止的包导入
	if err := checkForbiddenImports(code); err != nil {
		return err
	}

	// 检查危险关键字
	dangerousPatterns := []string{
		"syscall",
		"os.exec",
		"os.Remove",
		"os.RemoveAll",
		"ioutil.WriteFile",
		"os.OpenFile",
		"net.Listen",
		"net.Dial",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(code, pattern) {
			return fmt.Errorf("forbidden code pattern detected: %s", pattern)
		}
	}

	return nil
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
	// 获取客户端 IP
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	// 检查速率限制
	if !rateLimiter.Allow(ip) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证代码安全性
	if err := validateCode(req.Code); err != nil {
		resp := CodeResponse{Error: fmt.Sprintf("Security violation: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	output, err := executeCodeInDocker(req.Code)
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
	// 获取客户端 IP
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	// 检查速率限制
	if !rateLimiter.Allow(ip) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证代码安全性
	if err := validateCode(req.Code); err != nil {
		resp := FormatResponse{Error: fmt.Sprintf("Security violation: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
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

// executeCodeInDocker 在 Docker 容器中执行代码
func executeCodeInDocker(code string) (string, error) {
	// 创建唯一的容器名
	containerName := fmt.Sprintf("%s%d", containerPrefix, time.Now().UnixNano())

	// 创建上下文（用于超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout*time.Second)
	defer cancel()

	// 准备 Docker 运行命令
	args := []string{
		"run",
		"--name", containerName,
		"--rm",              // 运行完成后自动删除容器
		"--network", "none", // 禁用网络
		"--memory", "100m", // 限制内存使用
		"--cpus", "0.5", // 限制 CPU 使用
		"--pids-limit", "20", // 限制进程数
		"--ulimit", "nofile=64:64", // 限制文件描述符
		dockerImage,
		code,
	}

	// 运行容器
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 确保容器被清理
	cleanup := exec.Command("docker", "rm", "-f", containerName)
	cleanup.Run()

	// 处理执行结果
	output := stdout.String()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timeout after %d seconds", execTimeout)
		}
		errStr := stderr.String()
		if errStr != "" {
			return "", fmt.Errorf("execution error: %s", errStr)
		}
		return "", fmt.Errorf("execution error: %v", err)
	}

	// 检查输出长度
	if len(output) > maxOutputLength {
		return "", fmt.Errorf("output exceeds maximum allowed length (%d bytes)", maxOutputLength)
	}

	return output, nil
}
