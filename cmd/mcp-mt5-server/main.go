package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lukeware/kayron-ai/internal/config"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	"github.com/lukeware/kayron-ai/internal/services/daemon"
	"github.com/lukeware/kayron-ai/internal/services/health"
	"github.com/lukeware/kayron-ai/internal/services/mcp"
	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// MCPServer represents the MCP server
type MCPServer struct {
	logger   *logger.Logger
	daemon   *daemon.Daemon
	queue    *models.Queue
	mt5Client *mt5.Client

	// Service handlers
	accountInfoTool      *mcp.AccountInfoTool
	quoteTool            *mcp.QuoteTool
	placeOrderTool       *mcp.PlaceOrderTool
	closePositionTool    *mcp.ClosePositionTool
	ordersListTool       *mcp.OrdersListTool
	candleTool           *mcp.CandleTool
	modifyOrderTool      *mcp.ModifyOrderTool
	pendingOrderTool     *mcp.PendingOrderTool
	symbolPropertiesTool *mcp.SymbolPropertiesTool
	marginCalculatorTool *mcp.MarginCalculatorTool
}

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  interface{}   `json:"params"`
	ID      interface{}   `json:"id"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolRegistry holds all available MCP tools
type ToolRegistry struct {
	Tools map[string]func(interface{}) (interface{}, error)
}

// NewMCPServer creates a new MCP server
func NewMCPServer(cfg *config.Config) *MCPServer {
	queue, err := models.NewQueue("./mt5_queue.db")
	if err != nil {
		// For testing without CGO: skip queue, still functional for basic testing
		logger := logger.New("Main")
		logger.Info(fmt.Sprintf("Warning: queue initialization failed (requires CGO): %v. Proceeding without persistence.", err))
		queue = nil
	}

	mt5Client := mt5.NewClient(
		fmt.Sprintf("http://%s:8228", cfg.MT5Server),
		cfg.MT5Login,
		cfg.MT5Password,
		cfg.MT5Timeout,
	)

	grpcDaemon, err := daemon.NewDaemon(cfg.GRPCPort, mt5Client, queue)
	if err != nil {
		log.Fatalf("Failed to create daemon: %v", err)
	}

	// Initialize MT5 services
	accountService := mt5.NewAccountService(mt5Client)
	quoteService := mt5.NewQuoteService(mt5Client)
	orderService := mt5.NewOrderService(mt5Client)
	positionService := mt5.NewPositionService(mt5Client)
	ordersService := mt5.NewOrdersService(mt5Client)
	candleService := mt5.NewCandleService(mt5Client)
	modifyOrderService := mt5.NewModifyOrderService(mt5Client)
	pendingOrderService := mt5.NewPendingOrderService(mt5Client)
	symbolPropertiesService := mt5.NewSymbolPropertiesService(mt5Client)
	marginCalculatorService := mt5.NewMarginCalculatorService(mt5Client)

	// Initialize daemon services
	accountHandler := daemon.NewAccountServiceHandler(accountService)
	quoteHandler := daemon.NewQuoteServiceHandler(quoteService)
	orderHandler := daemon.NewOrderServiceHandler(orderService, queue)
	positionHandler := daemon.NewPositionServiceHandler(positionService)
	ordersHandler := daemon.NewOrdersServiceHandler(ordersService)
	candleHandler := daemon.NewCandleServiceHandler(candleService)
	modifyOrderHandler := daemon.NewModifyOrderServiceHandler(modifyOrderService)
	pendingOrderHandler := daemon.NewPendingOrderServiceHandler(pendingOrderService)
	symbolPropertiesHandler := daemon.NewSymbolPropertiesServiceHandler(symbolPropertiesService)
	marginCalculatorHandler := daemon.NewMarginCalculatorServiceHandler(marginCalculatorService)

	// Initialize MCP tools
	accountInfoTool := mcp.NewAccountInfoTool(accountHandler)
	quoteTool := mcp.NewQuoteTool(quoteHandler)
	placeOrderTool := mcp.NewPlaceOrderTool(orderHandler)
	closePositionTool := mcp.NewClosePositionTool(positionHandler)
	ordersListTool := mcp.NewOrdersListTool(ordersHandler)
	candleTool := mcp.NewCandleTool(candleHandler)
	modifyOrderTool := mcp.NewModifyOrderTool(modifyOrderHandler)
	pendingOrderTool := mcp.NewPendingOrderTool(pendingOrderHandler)
	symbolPropertiesTool := mcp.NewSymbolPropertiesTool(symbolPropertiesHandler)
	marginCalculatorTool := mcp.NewMarginCalculatorTool(marginCalculatorHandler)

	return &MCPServer{
		logger:               logger.New("MCPServer"),
		daemon:               grpcDaemon,
		queue:                queue,
		mt5Client:            mt5Client,
		accountInfoTool:      accountInfoTool,
		quoteTool:            quoteTool,
		placeOrderTool:       placeOrderTool,
		closePositionTool:    closePositionTool,
		ordersListTool:       ordersListTool,
		candleTool:           candleTool,
		modifyOrderTool:      modifyOrderTool,
		pendingOrderTool:     pendingOrderTool,
		symbolPropertiesTool: symbolPropertiesTool,
		marginCalculatorTool: marginCalculatorTool,
	}
}

// Start starts the MCP server
func (s *MCPServer) Start(cfg *config.Config) {
	// Start gRPC daemon in background
	go func() {
		if err := s.daemon.Start(); err != nil {
			s.logger.Error("Daemon failed to start", err)
		}
	}()

	// Setup HTTP server for JSON-RPC and health checks
	mux := http.NewServeMux()

	// JSON-RPC 2.0 endpoint
	mux.HandleFunc("/rpc", s.handleRPC)

	// Health check endpoint
	healthHandler := health.NewHandler(s.queue)
	mux.Handle("/health", healthHandler)

	// Start HTTP server
	httpAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	s.logger.Info(fmt.Sprintf("HTTP server listening on %s", httpAddr))

	if err := http.ListenAndServe(httpAddr, mux); err != nil && err != http.ErrServerClosed {
		s.logger.Error("HTTP server error", err)
	}
}

// handleRPC handles JSON-RPC 2.0 requests
func (s *MCPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, -32700, "Parse error")
		return
	}

	// Validate JSON-RPC format
	if req.Jsonrpc != "2.0" {
		respondWithError(w, -32600, "Invalid Request")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Register available tools
	registry := &ToolRegistry{
		Tools: map[string]func(interface{}) (interface{}, error){
			"account-info":      s.handleAccountInfo,
			"quote":             s.handleQuote,
			"place-order":       s.handlePlaceOrder,
			"close-position":    s.handleClosePosition,
			"orders-list":       s.handleOrdersList,
			"get-candles":       s.handleGetCandles,
			"modify-order":      s.handleModifyOrder,
			"pending-order-details": s.handlePendingOrderDetails,
			"symbol-properties":     s.handleSymbolProperties,
			"margin-calculator":     s.handleMarginCalculator,
		},
	}

	// Find and execute tool
	if tool, exists := registry.Tools[req.Method]; exists {
		result, err := tool(req.Params)
		if err != nil {
			respondError(w, req.ID, -32603, err.Error())
			return
		}
		respondSuccess(w, req.ID, result)
		return
	}

	respondError(w, req.ID, -32601, "Method not found")
}

// Tool implementations
func (s *MCPServer) handleAccountInfo(params interface{}) (interface{}, error) {
	return s.accountInfoTool.Execute(params)
}

func (s *MCPServer) handleQuote(params interface{}) (interface{}, error) {
	return s.quoteTool.Execute(params)
}

func (s *MCPServer) handlePlaceOrder(params interface{}) (interface{}, error) {
	return s.placeOrderTool.Execute(params)
}

func (s *MCPServer) handleClosePosition(params interface{}) (interface{}, error) {
	return s.closePositionTool.Execute(params)
}

func (s *MCPServer) handleOrdersList(params interface{}) (interface{}, error) {
	return s.ordersListTool.Execute(params)
}

func (s *MCPServer) handleGetCandles(params interface{}) (interface{}, error) {
	return s.candleTool.Execute(params)
}

func (s *MCPServer) handleModifyOrder(params interface{}) (interface{}, error) {
	return s.modifyOrderTool.Execute(params)
}

func (s *MCPServer) handlePendingOrderDetails(params interface{}) (interface{}, error) {
	return s.pendingOrderTool.Execute(params)
}

func (s *MCPServer) handleSymbolProperties(params interface{}) (interface{}, error) {
	return s.symbolPropertiesTool.Execute(params)
}

func (s *MCPServer) handleMarginCalculator(params interface{}) (interface{}, error) {
	return s.marginCalculatorTool.Execute(params)
}

// Response helpers
func respondSuccess(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := MCPResponse{
		Jsonrpc: "2.0",
		Result:  result,
		ID:      id,
	}
	json.NewEncoder(w).Encode(resp)
}

func respondError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := MCPResponse{
		Jsonrpc: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	json.NewEncoder(w).Encode(resp)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	resp := MCPResponse{
		Jsonrpc: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	cfg := config.Load()

	server := NewMCPServer(cfg)
	logger := logger.New("Main")

	logger.Info("Starting MT5 MCP Server")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		server.daemon.Stop()
		if server.queue != nil {
			server.queue.Close()
		}
		os.Exit(0)
	}()

	server.Start(cfg)
}
