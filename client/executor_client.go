package client

import (
	"context"

	"github.com/hanfei1991/microcosm/pb"
	"github.com/hanfei1991/microcosm/pkg/errors"
	"github.com/hanfei1991/microcosm/test"
	"github.com/hanfei1991/microcosm/test/mock"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type ExecutorClient interface {
	Send(context.Context, *ExecutorRequest) (*ExecutorResponse, error)
}

type closeableConnIface interface {
	Close() error
}

type executorClient struct {
	conn   closeableConnIface
	client pb.ExecutorClient
}

func (c *executorClient) Send(ctx context.Context, req *ExecutorRequest) (*ExecutorResponse, error) {
	resp := &ExecutorResponse{}
	var err error
	switch req.Cmd {
	case CmdSubmitBatchTasks:
		resp.Resp, err = c.client.SubmitBatchTasks(ctx, req.SubmitBatchTasks())
	case CmdCancelBatchTasks:
		resp.Resp, err = c.client.CancelBatchTasks(ctx, req.CancelBatchTasks())
	case CmdPauseBatchTasks:
		resp.Resp, err = c.client.PauseBatchTasks(ctx, req.PauseBatchTasks())
	case CmdDispatchTask:
		resp.Resp, err = c.client.DispatchTask(ctx, req.DispatchTask())
	}
	if err != nil {
		log.L().Logger.Error("send req meet error", zap.Error(err))
	}
	return resp, err
}

func newExecutorClientForTest(addr string) (*executorClient, error) {
	conn, err := mock.Dial(addr)
	if err != nil {
		return nil, errors.ErrGrpcBuildConn.GenWithStackByArgs(addr)
	}
	return &executorClient{
		conn:   conn,
		client: mock.NewExecutorClient(conn),
	}, nil
}

func newExecutorClient(addr string) (*executorClient, error) {
	if test.GetGlobalTestFlag() {
		return newExecutorClientForTest(addr)
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.DefaultConfig}))
	if err != nil {
		return nil, errors.ErrGrpcBuildConn.GenWithStackByArgs(addr)
	}
	return &executorClient{
		conn:   conn,
		client: pb.NewExecutorClient(conn),
	}, nil
}

type CmdType uint16

const (
	CmdSubmitBatchTasks CmdType = 1 + iota
	CmdCancelBatchTasks
	CmdPauseBatchTasks
	CmdDispatchTask
)

type ExecutorRequest struct {
	Cmd CmdType
	Req interface{}
}

func (e *ExecutorRequest) SubmitBatchTasks() *pb.SubmitBatchTasksRequest {
	return e.Req.(*pb.SubmitBatchTasksRequest)
}

func (e *ExecutorRequest) CancelBatchTasks() *pb.CancelBatchTasksRequest {
	return e.Req.(*pb.CancelBatchTasksRequest)
}

func (e *ExecutorRequest) PauseBatchTasks() *pb.PauseBatchTasksRequest {
	return e.Req.(*pb.PauseBatchTasksRequest)
}

func (e *ExecutorRequest) DispatchTask() *pb.DispatchTaskRequest {
	return e.Req.(*pb.DispatchTaskRequest)
}

type ExecutorResponse struct {
	Resp interface{}
}
