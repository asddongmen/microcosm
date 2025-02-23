package p2p

import (
	"context"
	"time"

	"github.com/pingcap/errors"
	cerror "github.com/pingcap/tiflow/pkg/errors"
	p2pImpl "github.com/pingcap/tiflow/pkg/p2p"
	"github.com/pingcap/tiflow/pkg/security"
)

// MessageSender is used to send a message of a given topic to a given node.
type MessageSender interface {
	// TODO investigate whether we need to implement a barrier mechanism

	// SendToNode sends a message to a given node. Returns whether it is successful and a possible error.
	// A `would-block` error will not be returned. (false, nil) would be returned instead.
	SendToNode(ctx context.Context, targetNodeID NodeID, topic Topic, message interface{}) (bool, error)
}

type messageSenderImpl struct {
	router MessageRouter
}

// NewMessageSender returns a new message sender.
func NewMessageSender(router MessageRouter) MessageSender {
	return &messageSenderImpl{router: router}
}

func (m *messageSenderImpl) SendToNode(ctx context.Context, targetNodeID NodeID, topic Topic, message interface{}) (bool, error) {
	client := m.router.GetClient(targetNodeID)
	if client == nil {
		return false, nil
	}

	_, err := client.TrySendMessage(ctx, topic, message)
	if err != nil {
		if cerror.ErrPeerMessageSendTryAgain.Equal(err) {
			return false, nil
		}
		return false, errors.Trace(err)
	}
	return true, nil
}

type MessageRouter = p2pImpl.MessageRouter

var defaultClientConfig = &p2pImpl.MessageClientConfig{
	SendChannelSize:         128,
	BatchSendInterval:       100 * time.Millisecond, // essentially disables flushing
	MaxBatchBytes:           8 * 1024 * 1024,        // 8MB
	MaxBatchCount:           4096,
	RetryRateLimitPerSecond: 1.0,      // once per second
	ClientVersion:           "v5.4.0", // a fake version
}

func NewMessageRouter(nodeID NodeID, advertisedAddr string) MessageRouter {
	config := *defaultClientConfig // copy
	config.AdvertisedAddr = advertisedAddr
	return p2pImpl.NewMessageRouter(
		nodeID,
		&security.Credential{ /* TLS not supported for now */ },
		&config,
	)
}
