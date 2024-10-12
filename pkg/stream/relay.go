package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-flv"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

type Publisher interface {
	Connect(context.Context, string) error
	Publish(io.Reader) error
	Close() error
}

type Option func(*Relayer)

func WithHTTPClient(c *http.Client) Option {
	return func(r *Relayer) {
		r.httpClient = c
	}
}

func WithOnRequest(fn OnRequestFunc) Option {
	return func(r *Relayer) {
		r.onRequest = fn
	}
}

type Relayer struct {
	src        string
	dst        string
	httpClient *http.Client
	onRequest  OnRequestFunc
	player     *HLSPlayer
	publisher  *RTMPPublisher
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

func NewRelayer(sourceURL, destinationURL string, opts ...Option) *Relayer {
	r := &Relayer{
		src:        strings.TrimSpace(sourceURL),
		dst:        strings.TrimSpace(destinationURL),
		httpClient: http.DefaultClient,
		onRequest:  func(_ *http.Request) {},
	}
	for _, o := range opts {
		o(r)
	}

	r.ctx, r.ctxCancel = context.WithCancel(context.Background())
	return r
}

func (r *Relayer) Run() error {
	if r.src == "" {
		return errors.New("source URL not set")
	}

	if r.publisher == nil {
		u, err := url.Parse(r.dst)
		if err != nil {
			return err
		}
		switch u.Scheme {
		case "rtmp", "rtmps":
			r.publisher = NewRTMPPublisher()
		case "srt":
		default:
			return fmt.Errorf("unsupported destination URL: %s", r.dst)
		}
	}

	err := r.publisher.Start(r.ctx, r.ctxCancel, r.dst)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(r.ctx, http.MethodGet, r.src, nil)
	if err != nil {
		slog.Error("Error creating request", "err", err)
		return err
	}

	r.onRequest(req)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		slog.Error("Error executing request", "error", err)
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	slog.Debug("Request completed", "final_url", resp.Request.URL.String(), "status", resp.Status, "content-type", contentType, "content-length", resp.ContentLength)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// read probe
	var pt string
	if strings.Contains(contentType, "application/vnd.apple.mpegurl") || strings.Contains(contentType, "application/x-mpegurl") {
		pt = "hls"
	}
	if strings.Contains(contentType, "video/x-flv") || strings.Contains(contentType, "video/x-p2p") {
		pt = "flv"
	}

	// Fallback to URL-based inference if Content-Type is inconclusive
	if pt == "" {
		pt = InferStreamType(r.src)
	}

	if pt == "hls" {
		demuxer := mpeg2.NewTSDemuxer()
		demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
			switch cid {
			case mpeg2.TS_STREAM_AAC:
				r.publisher.Cli.WriteAudio(codec.CODECID_AUDIO_AAC, frame, uint32(pts), uint32(dts))
			case mpeg2.TS_STREAM_H264:
				r.publisher.Cli.WriteVideo(codec.CODECID_VIDEO_H264, frame, uint32(pts), uint32(dts))
			case mpeg2.TS_STREAM_H265:
				r.publisher.Cli.WriteVideo(codec.CODECID_VIDEO_H265, frame, uint32(pts), uint32(dts))
			}
		}

		r.player = &HLSPlayer{
			URI:                 resp.Request.URL.String(),
			HTTPClient:          r.httpClient,
			OnRequest:           r.onRequest,
			OnSegmentDownloaded: func(b []byte) { demuxer.Input(bytes.NewReader(b)) },
			ctx:                 r.ctx,
			ctxCancel:           r.ctxCancel,
		}

		return r.player.Start()
	}

	if pt == "flv" {
		flvReader := flv.CreateFlvReader()
		flvReader.OnFrame = func(cid codec.CodecID, frame []byte, pts, dts uint32) {
			r.publisher.Cli.WriteFrame(cid, frame, pts, dts)
		}

		buf := make([]byte, 4096)
		for {
			select {
			case <-r.ctx.Done():
				slog.Info("Context cancelled, stopping FLV stream processing", "url", r.src)
				return nil
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					slog.Error("Error reading FLV stream", "err", err, "url", r.src)
					return err
				}
				if n > 0 {
					flvReader.Input(buf[:n])
				}
			}
		}
	}

	return fmt.Errorf("unsupported content type: %s", contentType)
}

func (r *Relayer) Stop() {
	r.ctxCancel()
}

func InferStreamType(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("Error parsing URL", "url", rawURL, "err", err)
		return ""
	}
	if strings.HasSuffix(u.Path, ".m3u8") {
		return "hls"
	}
	if strings.HasSuffix(u.Path, ".flv") || strings.HasSuffix(u.Path, ".xs") {
		return "flv"
	}
	return ""
}
