package stream

import (
	"context"
	"log/slog"
	"net"
	"net/url"

	"github.com/yapingcat/gomedia/go-rtmp"
)

type RTMPPublisher struct {
	Cli *rtmp.RtmpClient
}

func NewRTMPPublisher() *RTMPPublisher {
	return &RTMPPublisher{}
}

func (p *RTMPPublisher) Start(ctx context.Context, cancel context.CancelFunc, address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return err
	}

	host := u.Host
	if u.Port() == "" {
		host += ":1935"
	}

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}

	isReady := make(chan struct{})

	cli := rtmp.NewRtmpClient(rtmp.WithComplexHandshake(), rtmp.WithEnablePublish())
	cli.OnStateChange(func(newState rtmp.RtmpState) {
		if newState == rtmp.STATE_RTMP_PUBLISH_START {
			slog.Info("ready for publish", "url", address)
			close(isReady)
		}
	})
	cli.SetOutput(func(data []byte) error {
		_, err := conn.Write(data)
		return err
	})
	p.Cli = cli

	go func() {
		defer conn.Close()

		cli.Start(address)
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				slog.Info("RTMP publishing stopped", "err", ctx.Err(), "url", address)
				return
			default:
				n, err := conn.Read(buf)
				if err != nil {
					cancel() // Cancel the context, notifying other goroutines
					slog.Error("RTMP publishing read error", "err", err, "url", address)
					return
				}
				cli.Input(buf[:n])
			}
		}
	}()

	// waiting for ready or timeout
	select {
	case <-isReady:
		slog.Debug("RTMP publisher is ready to publish")
	case <-ctx.Done():
		slog.Error("RTMP publisher start timeout", "error", ctx.Err())
		return ctx.Err()
	}

	return nil
}
