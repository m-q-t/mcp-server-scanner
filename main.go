package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mcp-server-scanner/pkg/listtools"
)

func main() {
	baseURL := flag.String("url", "", "Base URL for MCP endpoint, e.g. http://localhost:8080")
	timeout := flag.Int("timeout", 5, "Timeout in seconds")

	flag.Parse()

	if *baseURL == "" {
		log.Fatal("URL parameter is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	client := listtools.NewClient(*baseURL)
	defer client.Close()

	tools, err := listtools.FetchToolsResponse(ctx, client)
	if err != nil {
		log.Fatalf("Failed to fetch tools: %v", err)
	}

	log.Printf("Tools: %v", tools)
}
