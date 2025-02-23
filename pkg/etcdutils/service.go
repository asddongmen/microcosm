package etcdutils

import (
	"context"
	"net/http"
	"time"

	"github.com/hanfei1991/microcosm/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/embed"
	"google.golang.org/grpc"
)

// StartEtcd starts an embedded etcd server.
func StartEtcd(etcdCfg *embed.Config,
	gRPCSvr func(*grpc.Server),
	httpHandles map[string]http.Handler,
	startTimeout time.Duration,
) (*embed.Etcd, error) {
	// attach extra gRPC and HTTP server
	if gRPCSvr != nil {
		etcdCfg.ServiceRegister = gRPCSvr
	}
	if httpHandles != nil {
		etcdCfg.UserHandlers = httpHandles
	}

	e, err := embed.StartEtcd(etcdCfg)
	if err != nil {
		return nil, errors.Wrap(errors.ErrMasterStartEmbedEtcdFail, err)
	}

	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(startTimeout):
		// if fail to startup, the etcd server may be still blocking in
		// https://github.com/etcd-io/etcd/blob/3cf2f69b5738fb702ba1a935590f36b52b18979b/embed/serve.go#L92
		// then `e.Close` will block in
		// https://github.com/etcd-io/etcd/blob/3cf2f69b5738fb702ba1a935590f36b52b18979b/embed/etcd.go#L377
		// because `close(sctx.serversC)` has not been called in
		// https://github.com/etcd-io/etcd/blob/3cf2f69b5738fb702ba1a935590f36b52b18979b/embed/serve.go#L200.
		// so for `ReadyNotify` timeout, we choose to only call `e.Server.Stop()` now,
		// and we should exit the DM-master process after returned with error from this function.
		e.Server.Stop()
		return nil, errors.ErrMasterStartEmbedEtcdFail.GenWithStack("start embed etcd timeout %v", startTimeout)
	}
	return e, nil
}

// GetLeader returns the campaign value and revision based on given campaign key
func GetLeader(ctx context.Context, cli *clientv3.Client, campKey string) (
	key []byte, val []byte, rev int64, err error,
) {
	opts := append([]clientv3.OpOption{clientv3.WithPrefix()}, clientv3.WithFirstCreate()...)
	resp, err := cli.Get(ctx, campKey, opts...)
	if err != nil {
		err = errors.Wrap(errors.ErrEtcdAPIError, err)
		return
	}
	if len(resp.Kvs) == 0 {
		err = errors.ErrMasterNoLeader.GenWithStackByArgs()
		return
	}
	key = resp.Kvs[0].Key
	val = resp.Kvs[0].Value
	rev = resp.Header.Revision
	return
}
