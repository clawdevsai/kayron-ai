package daemon

import (
	"context"
	"fmt"
	"net"

	pb "github.com/lukeware/kayron-ai/api/mt5"
	"github.com/lukeware/kayron-ai/internal/logger"
	"github.com/lukeware/kayron-ai/internal/models"
	mt5client "github.com/lukeware/kayron-ai/internal/services/mt5"
	"google.golang.org/grpc"
)

// Daemon manages the gRPC server and service implementations
type Daemon struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
	mt5Client  *mt5client.Client
	queue      *models.Queue
	pb.UnimplementedMT5ServiceServer
}

// NewDaemon creates a new gRPC daemon
func NewDaemon(port int, mt5Client *mt5client.Client, queue *models.Queue) (*Daemon, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	return &Daemon{
		grpcServer: grpc.NewServer(),
		listener:   listener,
		logger:     logger.New("Daemon"),
		mt5Client:  mt5Client,
		queue:      queue,
	}, nil
}

// Start starts the gRPC server
func (d *Daemon) Start() error {
	// Register service implementation
	pb.RegisterMT5ServiceServer(d.grpcServer, d)

	d.logger.Info("gRPC daemon starting on localhost:50051")

	if err := d.grpcServer.Serve(d.listener); err != nil {
		d.logger.Error("gRPC server error", err)
		return err
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (d *Daemon) Stop() {
	d.logger.Info("Shutting down gRPC daemon")
	d.grpcServer.GracefulStop()
}

// GetAccountInfo implements the gRPC GetAccountInfo method
func (d *Daemon) GetAccountInfo(ctx context.Context, req *pb.AccountInfoRequest) (*pb.AccountInfoResponse, error) {
	// TODO: Implement after user stories
	return nil, nil
}

// GetQuote implements the gRPC GetQuote method
func (d *Daemon) GetQuote(ctx context.Context, req *pb.QuoteRequest) (*pb.QuoteResponse, error) {
	// TODO: Implement after user stories
	return nil, nil
}

// PlaceOrder implements the gRPC PlaceOrder method
func (d *Daemon) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	// TODO: Implement after user stories
	return nil, nil
}

// ClosePosition implements the gRPC ClosePosition method
func (d *Daemon) ClosePosition(ctx context.Context, req *pb.ClosePositionRequest) (*pb.ClosePositionResponse, error) {
	// TODO: Implement after user stories
	return nil, nil
}

// ListOrders implements the gRPC ListOrders method
func (d *Daemon) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	// TODO: Implement after user stories
	return nil, nil
}
