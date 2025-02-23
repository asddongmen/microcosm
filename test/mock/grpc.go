package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/hanfei1991/microcosm/pb"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var container *grpcContainer

func init() {
	ResetGrpcCtx()
}

type grpcContainer struct {
	mu      sync.Mutex
	servers map[string]GrpcServer
}

// GrpcServer implements a mock grpc server.
type GrpcServer interface {
	dial() (Conn, error)
	Stop()
}

type masterServer struct {
	*baseServer
	pb.MasterServer
}

func (s *masterServer) dial() (Conn, error) {
	return &masterServerConn{s}, nil
}

type masterServerConn struct {
	server *masterServer
}

func (s *masterServerConn) Close() error {
	return nil
}

func (s *masterServerConn) sendRequest(ctx context.Context, req interface{}) (interface{}, error) {
	switch x := req.(type) {
	case *pb.RegisterExecutorRequest:
		return s.server.RegisterExecutor(ctx, x)
	case *pb.PauseJobRequest:
		return s.server.PauseJob(ctx, x)
	case *pb.SubmitJobRequest:
		return s.server.SubmitJob(ctx, x)
	case *pb.HeartbeatRequest:
		return s.server.Heartbeat(ctx, x)
	case *pb.TaskSchedulerRequest:
		return s.server.ScheduleTask(ctx, x)
	case *pb.CancelJobRequest:
		return s.server.CancelJob(ctx, x)
	}
	return nil, errors.New("unknown request")
}

type masterServerClient struct {
	conn Conn
}

func (c *masterServerClient) RegisterExecutor(ctx context.Context, req *pb.RegisterExecutorRequest, opts ...grpc.CallOption) (*pb.RegisterExecutorResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.RegisterExecutorResponse), err
}

func (c *masterServerClient) PauseJob(ctx context.Context, req *pb.PauseJobRequest, opts ...grpc.CallOption) (*pb.PauseJobResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.PauseJobResponse), err
}

func (c *masterServerClient) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest, opts ...grpc.CallOption) (*pb.SubmitJobResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.SubmitJobResponse), err
}

func (c *masterServerClient) CancelJob(ctx context.Context, req *pb.CancelJobRequest, opts ...grpc.CallOption) (*pb.CancelJobResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.CancelJobResponse), err
}

func (c *masterServerClient) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest, opts ...grpc.CallOption) (*pb.HeartbeatResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.HeartbeatResponse), err
}

func (c *masterServerClient) ScheduleTask(ctx context.Context, req *pb.TaskSchedulerRequest, opts ...grpc.CallOption) (*pb.TaskSchedulerResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.TaskSchedulerResponse), nil
}

func (c *masterServerClient) RegisterMetaStore(
	ctx context.Context, req *pb.RegisterMetaStoreRequest, opts ...grpc.CallOption,
) (*pb.RegisterMetaStoreResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RegisterMetaStoreResponse), nil
}

func (c *masterServerClient) QueryMetaStore(
	ctx context.Context, req *pb.QueryMetaStoreRequest, opts ...grpc.CallOption,
) (*pb.QueryMetaStoreResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.QueryMetaStoreResponse), nil
}

func (c *masterServerClient) ReportExecutorWorkload(
	ctx context.Context, req *pb.ExecWorkloadRequest, opts ...grpc.CallOption,
) (*pb.ExecWorkloadResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ExecWorkloadResponse), nil
}

func NewMasterClient(conn Conn) pb.MasterClient {
	return &masterServerClient{conn}
}

type executorServer struct {
	*baseServer
	pb.ExecutorServer
}

type executorServerConn struct {
	server *executorServer
}

type executorClient struct {
	conn Conn
}

type baseServer struct {
	addr string
}

func (s *baseServer) Stop() {
	container.mu.Lock()
	defer container.mu.Unlock()
	_, ok := container.servers[s.addr]
	log.L().Logger.Info("server is cancelled", zap.String("ip", s.addr))
	if ok {
		delete(container.servers, s.addr)
	}
}

func (s *executorServer) dial() (Conn, error) {
	return &executorServerConn{s}, nil
}

func (c *executorClient) SubmitBatchTasks(ctx context.Context, req *pb.SubmitBatchTasksRequest, opts ...grpc.CallOption) (*pb.SubmitBatchTasksResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.SubmitBatchTasksResponse), err
}

func (c *executorClient) CancelBatchTasks(ctx context.Context, req *pb.CancelBatchTasksRequest, opts ...grpc.CallOption) (*pb.CancelBatchTasksResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.CancelBatchTasksResponse), err
}

func (c *executorClient) PauseBatchTasks(ctx context.Context, req *pb.PauseBatchTasksRequest, opts ...grpc.CallOption) (*pb.PauseBatchTasksResponse, error) {
	resp, err := c.conn.sendRequest(ctx, req)
	return resp.(*pb.PauseBatchTasksResponse), err
}

func (c *executorClient) DispatchTask(ctx context.Context, in *pb.DispatchTaskRequest, opts ...grpc.CallOption) (*pb.DispatchTaskResponse, error) {
	panic("implement me")
}

func (s *executorServerConn) Close() error {
	return nil
}

func NewExecutorClient(conn Conn) pb.ExecutorClient {
	return &executorClient{conn}
}

func (s *executorServerConn) sendRequest(ctx context.Context, req interface{}) (interface{}, error) {
	switch x := req.(type) {
	case *pb.SubmitBatchTasksRequest:
		return s.server.SubmitBatchTasks(ctx, x)
	case *pb.CancelBatchTasksRequest:
		return s.server.CancelBatchTasks(ctx, x)
	case *pb.PauseBatchTasksRequest:
		return s.server.PauseBatchTasks(ctx, x)
	}
	return nil, errors.New("unknown request")
}

func Dial(addr string) (Conn, error) {
	container.mu.Lock()
	defer container.mu.Unlock()
	server, ok := container.servers[addr]
	if !ok {
		return nil, errors.New("no server found")
	}
	return server.dial()
}

// NewMasterServer creates a master grpc server and listened the address.
// We try to make things simple, so we design the "NewMasterServer" to register only one type of pb server.
func NewMasterServer(addr string, server pb.MasterServer) (GrpcServer, error) {
	container.mu.Lock()
	defer container.mu.Unlock()
	_, ok := container.servers[addr]
	if ok {
		return nil, errors.New("addr " + addr + " has been listened")
	}
	newServer := &masterServer{&baseServer{addr}, server}
	container.servers[addr] = newServer
	return newServer, nil
}

func NewExecutorServer(addr string, server pb.ExecutorServer) (GrpcServer, error) {
	container.mu.Lock()
	defer container.mu.Unlock()
	_, ok := container.servers[addr]
	if ok {
		return nil, errors.New("addr " + addr + " has been listened")
	}
	newServer := &executorServer{&baseServer{addr}, server}
	container.servers[addr] = newServer
	return newServer, nil
}

type Conn interface {
	Close() error
	sendRequest(ctx context.Context, req interface{}) (interface{}, error)
}

func ResetGrpcCtx() {
	container = &grpcContainer{
		servers: make(map[string]GrpcServer),
	}
}
