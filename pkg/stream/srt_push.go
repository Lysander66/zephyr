package stream

import (
	"context"
	"io"
)

type SRTPublisher struct {
}

func (S SRTPublisher) Connect(ctx context.Context, s string) error {
	panic("implement me")
}

func (S SRTPublisher) Publish(reader io.Reader) error {
	panic("implement me")
}

func (S SRTPublisher) Close() error {
	panic("implement me")
}
