package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/borankux/gops/internal/port"
	"github.com/borankux/gops/internal/process"
	"github.com/borankux/gops/internal/resource"
	"github.com/borankux/gops/internal/service"
	"github.com/borankux/gops/internal/window"
	"github.com/borankux/gops/pkg/types"
)

// Server represents the MCP server
type Server struct {
	port   int
	server *http.Server
}

// NewServer creates a new MCP server
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start starts the MCP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// MCP protocol endpoints with CORS support
	mux.HandleFunc("/mcp/v1/processes", s.corsMiddleware(s.handleProcesses))
	mux.HandleFunc("/mcp/v1/windows", s.corsMiddleware(s.handleWindows))
	mux.HandleFunc("/mcp/v1/ports", s.corsMiddleware(s.handlePorts))
	mux.HandleFunc("/mcp/v1/resource", s.corsMiddleware(s.handleResource))
	mux.HandleFunc("/mcp/v1/services", s.corsMiddleware(s.handleServices))
	mux.HandleFunc("/health", s.corsMiddleware(s.handleHealth))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Printf("🚀 MCP Server starting on port %d", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the MCP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	procs, err := process.GetUserApplications(ctx)
	if err != nil {
		s.sendError(w, err)
		return
	}

	response := types.ProcessesResponse{
		Processes: procs,
		Count:     len(procs),
	}

	s.sendJSON(w, response)
}

func (s *Server) handleWindows(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	windows, err := window.GetOpenWindows(ctx)
	if err != nil {
		s.sendError(w, err)
		return
	}

	response := types.WindowsResponse{
		Windows: windows,
		Count:   len(windows),
	}

	s.sendJSON(w, response)
}

func (s *Server) handlePorts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	portParam := r.URL.Query().Get("port")
	pidParam := r.URL.Query().Get("pid")

	var ports []types.PortInfo
	var err error

	if portParam != "" {
		portNum, parseErr := strconv.ParseUint(portParam, 10, 32)
		if parseErr != nil {
			s.sendError(w, fmt.Errorf("invalid port number: %w", parseErr))
			return
		}
		ports, err = port.GetPortInfoByPort(ctx, uint32(portNum))
	} else if pidParam != "" {
		pid, parseErr := strconv.ParseInt(pidParam, 10, 32)
		if parseErr != nil {
			s.sendError(w, fmt.Errorf("invalid PID: %w", parseErr))
			return
		}
		ports, err = port.GetPortsByPID(ctx, int32(pid))
	} else {
		ports, err = port.GetOpenPorts(ctx)
	}

	if err != nil {
		s.sendError(w, err)
		return
	}

	response := types.PortsResponse{
		Ports: ports,
		Count: len(ports),
	}

	s.sendJSON(w, response)
}

func (s *Server) handleResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	pidParam := r.URL.Query().Get("pid")
	if pidParam == "" {
		s.sendError(w, fmt.Errorf("pid parameter is required"))
		return
	}

	pid, err := strconv.ParseInt(pidParam, 10, 32)
	if err != nil {
		s.sendError(w, fmt.Errorf("invalid PID: %w", err))
		return
	}

	usage, err := resource.GetProcessResourceUsage(ctx, int32(pid))
	if err != nil {
		s.sendError(w, err)
		return
	}

	response := types.ResourceResponse{
		Usage: *usage,
	}

	s.sendJSON(w, response)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	services, err := service.GetServices(ctx)
	if err != nil {
		s.sendError(w, err)
		return
	}

	response := types.ServicesResponse{
		Services: services,
		Count:    len(services),
	}

	s.sendJSON(w, response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"status":"healthy"}`)
}

func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func (s *Server) sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	response := types.ErrorResponse{
		Error: err.Error(),
	}
	json.NewEncoder(w).Encode(response)
}

// corsMiddleware adds CORS headers to responses
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
