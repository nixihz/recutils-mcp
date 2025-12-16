package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/recutils-mcp/recutils-mcp/server"
)

func main() {
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
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	// Run server
	fmt.Println("Starting Recutils MCP Server...")
	fmt.Println("Press Ctrl+C to stop")

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
