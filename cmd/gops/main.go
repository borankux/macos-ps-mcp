package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/allintech/gops/internal/cli"
	"github.com/allintech/gops/internal/mcp"
)

func main() {
	var (
		// CLI flags
		processes = flag.Bool("processes", false, "List user applications")
		windows   = flag.Bool("windows", false, "List open windows")
		ports     = flag.Bool("ports", false, "List open ports")
		resource  = flag.Bool("resource", false, "Show resource usage for a process")
		services  = flag.Bool("services", false, "List system services")
		portFilter = flag.String("port", "", "Filter ports by port number")
		pid       = flag.String("pid", "", "Filter ports by PID or show resource usage")
		
		// MCP server flags
		serverMode = flag.Bool("server", false, "Start MCP server")
		serverPort = flag.Int("server-port", 8080, "MCP server port (default: 8080)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "🔧 gops - Process and System Information Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [mode] [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  CLI Mode (default):\n")
		fmt.Fprintf(os.Stderr, "    -processes              List all user applications\n")
		fmt.Fprintf(os.Stderr, "    -windows                 List open windows\n")
		fmt.Fprintf(os.Stderr, "    -ports                   List all open ports\n")
		fmt.Fprintf(os.Stderr, "    -ports -port 8080        Show info for port 8080\n")
		fmt.Fprintf(os.Stderr, "    -resource -pid 1234      Show resource usage for PID 1234\n")
		fmt.Fprintf(os.Stderr, "    -services                List system services\n\n")
		fmt.Fprintf(os.Stderr, "  MCP Server Mode:\n")
		fmt.Fprintf(os.Stderr, "    -server                  Start MCP server\n")
		fmt.Fprintf(os.Stderr, "    -server-port 8080        MCP server port (default: 8080)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -processes              List all user applications\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -server                 Start MCP server on port 8080\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -server -server-port 3000  Start MCP server on port 3000\n", os.Args[0])
	}

	flag.Parse()

	ctx := context.Background()

	// MCP Server Mode
	if *serverMode {
		server := mcp.NewServer(*serverPort)
		
		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		
		go func() {
			<-sigChan
			fmt.Println("\n🛑 Shutting down MCP server...")
			if err := server.Stop(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error stopping server: %v\n", err)
			}
			os.Exit(0)
		}()
		
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error starting MCP server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI Mode
	if *processes {
		if err := cli.DisplayProcesses(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *windows {
		if err := cli.DisplayWindows(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *ports {
		if err := cli.DisplayPorts(ctx, *portFilter, *pid); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *resource {
		if *pid == "" {
			fmt.Fprintf(os.Stderr, "❌ Error: -pid is required for -resource\n")
			os.Exit(1)
		}
		pidInt, err := strconv.ParseInt(*pid, 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: invalid PID: %v\n", err)
			os.Exit(1)
		}
		if err := cli.DisplayResourceUsage(ctx, int32(pidInt)); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *services {
		if err := cli.DisplayServices(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default: show help
	fmt.Println("🔧 gops - Process and System Information Tool\n")
	fmt.Println("Available commands:")
	fmt.Println("  -processes    List user applications")
	fmt.Println("  -windows      List open windows")
	fmt.Println("  -ports        List open ports")
	fmt.Println("  -resource     Show resource usage (requires -pid)")
	fmt.Println("  -services     List system services")
	fmt.Println("  -server       Start MCP server")
	fmt.Println("\nUse -help for more information")
}

