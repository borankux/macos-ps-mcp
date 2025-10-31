# gops - Process and System Information Tool

A powerful command-line tool and MCP (Model Context Protocol) server built in Go for listing processes, windows, ports, and system services with detailed resource usage information.

## Features

- 📱 **List User Applications**: Lists non-system background processes running on your system
- 🪟 **List Open Windows**: Displays all open windows with their associated processes
- 🌐 **List Open Ports**: Shows ports that are listening (open) with process information
- 📊 **Resource Usage**: Get detailed CPU and memory usage for specific processes
- ⚙️ **System Services**: List system services with their resource usage
- 🎨 **Beautiful CLI**: Formatted tables with emoji-rich output for command-line usage
- 📡 **MCP Server**: JSON API endpoints for integration with other tools

## Installation

```bash
go build -o gops ./cmd/gops
```

Or install directly:

```bash
go install ./cmd/gops
```

## Usage

### Command-Line Mode

#### List User Applications
```bash
./gops -processes
```

#### List Open Windows
```bash
./gops -windows
```

#### List Open Ports
```bash
# List all listening ports
./gops -ports

# Filter by port number
./gops -ports -port 8080

# Filter by PID
./gops -ports -pid 1234
```

#### Get Process Resource Usage
```bash
./gops -resource -pid 1234
```

#### List System Services
```bash
./gops -services
```

### MCP Server Mode

Start the MCP server:

```bash
# Default port (8080)
./gops -server

# Custom port
./gops -server -server-port 3000
```

#### API Endpoints

All endpoints return JSON responses:

- `GET /mcp/v1/processes` - List user applications
- `GET /mcp/v1/windows` - List open windows
- `GET /mcp/v1/ports?port=8080` - List open ports (optional: filter by port)
- `GET /mcp/v1/ports?pid=1234` - List ports by PID
- `GET /mcp/v1/resource?pid=1234` - Get resource usage for a process
- `GET /mcp/v1/services` - List system services
- `GET /health` - Health check endpoint

#### Example API Calls

```bash
# List processes
curl http://localhost:8080/mcp/v1/processes

# List ports
curl http://localhost:8080/mcp/v1/ports

# Get resource usage
curl http://localhost:8080/mcp/v1/resource?pid=1234

# List services
curl http://localhost:8080/mcp/v1/services
```

## Project Structure

```
gops/
├── cmd/
│   └── gops/
│       └── main.go          # Entry point with CLI and server modes
├── internal/
│   ├── cli/
│   │   └── cli.go           # CLI display functions with formatted tables
│   ├── mcp/
│   │   └── server.go        # MCP HTTP server implementation
│   ├── process/
│   │   └── process.go       # Process listing and filtering
│   ├── window/
│   │   └── window.go        # Window detection (macOS/Linux/Windows)
│   ├── port/
│   │   └── port.go          # Port listing and filtering
│   ├── resource/
│   │   └── resource.go      # CPU/Memory usage retrieval
│   ├── service/
│   │   └── service.go       # System service listing
│   └── utils/
│       └── format.go        # Human-readable formatting utilities
└── pkg/
    └── types/
        └── types.go         # Type definitions
```

## Platform Support

- ✅ **macOS**: Full support (uses osascript for windows, launchctl for services)
- ✅ **Linux**: Full support (uses wmctrl for windows, systemctl for services)
- ✅ **Windows**: Full support (uses PowerShell for windows and services)

## Requirements

- Go 1.21 or later
- On macOS: Requires AppleScript permissions for window detection
- On Linux: `wmctrl` package for window detection (optional)
- On Windows: PowerShell (included by default)

## Examples

### CLI Output Example

```
📱 User Applications

┌──────┬─────────────────────┬───────────────┬────────────────────────────────────┐
│ 🔢 PID│ 📛 Name              │ 👤 User       │ 📍 Path                            │
├──────┼─────────────────────┼───────────────┼────────────────────────────────────┤
│  1234│ Google Chrome       │ user          │ /Applications/Google Chrome.app   │
│  5678│ Visual Studio Code  │ user          │ /Applications/VS Code.app         │
└──────┴─────────────────────┴───────────────┴────────────────────────────────────┘
```

### JSON API Response Example

```json
{
  "processes": [
    {
      "pid": 1234,
      "name": "Google Chrome",
      "path": "/Applications/Google Chrome.app",
      "user": "user"
    }
  ],
  "count": 1
}
```

## License

MIT

