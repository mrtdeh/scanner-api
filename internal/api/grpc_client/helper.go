package grpc_client

import (
	"context"
)

func (r *Client) newCombinedContext(ctx context.Context) (context.Context, context.CancelFunc) {
	combinedCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-r.ctx.Done():
			cancel()
		case <-combinedCtx.Done():
			return
		}
	}()
	return combinedCtx, cancel
}
