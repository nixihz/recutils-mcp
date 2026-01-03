package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nixihz/recutils-mcp/server"
)

func getLogFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果获取失败，回退到当前目录
		return "logs/recutils-mcp.log"
	}
	return filepath.Join(homeDir, ".recutils-mcp", "recutils-mcp.log")
}

func initLogging() (*os.File, error) {
	logFilePath := getLogFilePath()

	// 确保日志目录存在
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 打开日志文件（追加模式）
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// 设置日志输出到文件和标准输出
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return logFile, nil
}

func main() {
	// 初始化日志
	logFile, err := initLogging()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Create context with graceful shutdown support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create MCP server
	srv := server.NewMCPServer()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	// Run server
	log.Println("Starting Recutils MCP Server...")

	if err := srv.Run(ctx); err != nil {
		log.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
