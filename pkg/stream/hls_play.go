package stream

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bluenviron/gohlslib/pkg/playlist"
)

const defaultLiveStartIndex = -3

type (
	OnRequestFunc            func(*http.Request)
	OnPlaylistDownloadedFunc func([]byte, playlist.Playlist)
	OnSegmentDownloadedFunc  func([]byte)
	FetchSegmentFunc         func(string, int) error
)

type HLSPlayer struct {
	// URI of the playlist
	URI string
	// Segment index to start live streams at (negative values are from the end)
	LiveStartIndex *int
	// HTTP client used for making requests
	HTTPClient *http.Client

	// Callbacks (all optional)

	// Called before each HTTP request
	OnRequest OnRequestFunc
	// Called after downloading a playlist
	OnPlaylistDownloaded OnPlaylistDownloadedFunc
	// Called after downloading a segment
	OnSegmentDownloaded OnSegmentDownloadedFunc
	// Custom function for fetching media segments
	FetchSegment FetchSegmentFunc

	// Private fields

	// Context for managing cancellation and timeouts
	ctx context.Context
	// Function to cancel the context
	ctxCancel context.CancelFunc
	// Parsed URL of the playlist
	playlistURL *url.URL
	// Timestamp of last playlist load (in milliseconds)
	lastLoadTimeMillis int64
	// Current media sequence number being processed
	curSeqNo int
	// Channel for reporting errors
	outErr chan error
}

func (c *HLSPlayer) Run() error {
	var err error
	c.playlistURL, err = url.Parse(c.URI)
	if err != nil {
		return err
	}

	if c.LiveStartIndex == nil {
		c.LiveStartIndex = intPtr(defaultLiveStartIndex)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	if c.OnRequest == nil {
		c.OnRequest = func(*http.Request) {}
	}

	if c.OnPlaylistDownloaded == nil {
		c.OnPlaylistDownloaded = func([]byte, playlist.Playlist) {}
	}
	if c.OnSegmentDownloaded == nil {
		c.OnSegmentDownloaded = func([]byte) {}
	}

	if c.FetchSegment == nil {
		c.FetchSegment = func(url string, _ int) error {
			b, err := fetch(c.ctx, c.HTTPClient, c.OnRequest, url)
			if err != nil {
				return err
			}
			c.OnSegmentDownloaded(b)
			return nil
		}
	}

	if c.ctx == nil {
		c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	}
	c.outErr = make(chan error, 1)

	go func() {
		c.outErr <- c.run()
	}()

	return nil
}

func (c *HLSPlayer) Wait() error {
	return <-c.outErr
}

func (c *HLSPlayer) Stop() {
	c.ctxCancel()
}

func (c *HLSPlayer) run() error {
	b, err := fetch(c.ctx, c.HTTPClient, c.OnRequest, c.playlistURL.String())
	if err != nil {
		return err
	}
	c.lastLoadTimeMillis = time.Now().UnixMilli()

	pl, err := playlist.Unmarshal(b)
	if err != nil {
		return err
	}

	c.OnPlaylistDownloaded(b, pl)

	var pls *playlist.Media

	switch plt := pl.(type) {
	case *playlist.Multivariant: // Master Playlist
		leadingPlaylist := pickLeadingPlaylist(plt.Variants)
		if leadingPlaylist == nil {
			return fmt.Errorf("no variants with supported codecs found")
		}

		u, err := AbsoluteURL(c.playlistURL, leadingPlaylist.URI)
		if err != nil {
			return err
		}
		c.playlistURL = u

		mediaPlaylist, err := c.fetchMediaPlaylist(u.String())
		if err != nil {
			return err
		}

		pls = mediaPlaylist

		if leadingPlaylist.Audio != "" {
			// TODO(Lysander)
		}
	case *playlist.Media: // Media Playlist
		pls = plt
	default:
		return fmt.Errorf("invalid playlist")
	}

	// Select the starting segments
	c.curSeqNo = selectCurSeqNo(*c.LiveStartIndex, pls)

	for i := c.curSeqNo - pls.MediaSequence; i < len(pls.Segments); i++ {
		if err := c.fetchMediaSegment(pls.Segments[i], i); err != nil {
			return err
		}
	}
	c.curSeqNo = pls.MediaSequence + max(len(pls.Segments)-1, 0)

	// Video on demand (VOD)
	if isVOD(pls) {
		return nil
	}

	// Live stream
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			pls, err := c.fetchMediaPlaylist(c.playlistURL.String())
			if err != nil {
				slog.Error("fetchMediaPlaylist", "err", err, "url", c.playlistURL.String())
				return err
			}

			reloadInterval := defaultReloadInterval(pls)

			// If there's still no more segments, switch to a reload interval of half the target duration.
			if pls.MediaSequence+len(pls.Segments)-1 <= c.curSeqNo {
				reloadInterval = time.Duration(max(pls.TargetDuration/2, 1)) * time.Second
			}

			timer.Reset(reloadInterval)

			for i, seg := range pls.Segments {
				if pls.MediaSequence+i > c.curSeqNo {
					if err := c.fetchMediaSegment(seg, i); err != nil {
						slog.Error("fetchMediaSegment", "err", err, "url", seg.URI)
					}
					c.curSeqNo = pls.MediaSequence + i
				}
			}

			if pls.Endlist {
				c.Stop()
				slog.Info("no more media segments")
			}
		case <-c.ctx.Done():
			slog.Info("reloadMediaPlaylist: quit", "url", c.playlistURL.String())
			return nil
		}
	}

}

func (c *HLSPlayer) fetchMediaPlaylist(url string) (*playlist.Media, error) {
	b, err := fetch(c.ctx, c.HTTPClient, c.OnRequest, url)
	if err != nil {
		return nil, err
	}
	c.lastLoadTimeMillis = time.Now().UnixMilli()

	pl, err := playlist.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	pls, ok := pl.(*playlist.Media)
	if !ok {
		return nil, fmt.Errorf("invalid media playlist")
	}

	c.OnPlaylistDownloaded(b, pls)

	return pls, nil
}

func (c *HLSPlayer) fetchMediaSegment(seg *playlist.MediaSegment, i int) error {
	u, err := AbsoluteURL(c.playlistURL, seg.URI)
	if err != nil {
		slog.Error("AbsoluteURL", "err", err, "url", seg.URI)
		return err
	}

	err = c.FetchSegment(u.String(), i)
	if err != nil {
		slog.Error("fetchMediaSegment", "err", err, "url", u.String())
		return err
	}

	return nil
}

/*
[spec 6.3.3](https://datatracker.ietf.org/doc/html/rfc8216#section-6.3.3)
If the EXT-X-ENDLIST tag is not present and the client intends to play the media normally, the client
SHOULD NOT choose a segment that starts less than three target durations from the end of the Playlist file.
*/
func selectCurSeqNo(liveStartIndex int, pls *playlist.Media) (seqNo int) {
	if !isVOD(pls) {
		// If this is a live stream, start live_start_index segments from the start or end
		if liveStartIndex < 0 {
			seqNo = pls.MediaSequence + max(liveStartIndex+len(pls.Segments), 0)
		} else {
			seqNo = pls.MediaSequence + min(liveStartIndex, len(pls.Segments)-1)
		}

		// If #EXT-X-START in playlist, need to recalculate
		if pls.Start != nil {
			// TODO(Lysander)
		}

		return seqNo
	}

	// Otherwise just start on the first segment
	return pls.MediaSequence
}

func defaultReloadInterval(pls *playlist.Media) time.Duration {
	if len(pls.Segments) > 0 {
		return pls.Segments[len(pls.Segments)-1].Duration
	}
	return time.Duration(pls.TargetDuration) * time.Second
}

func isVOD(pls *playlist.Media) bool {
	return pls.Endlist || (pls.PlaylistType != nil && *pls.PlaylistType == playlist.MediaPlaylistTypeVOD)
}

func intPtr(i int) *int {
	return &i
}

func fetch(ctx context.Context, httpClient *http.Client, onRequest OnRequestFunc, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	onRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func AbsoluteURL(base *url.URL, relative string) (*url.URL, error) {
	u, err := url.Parse(relative)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(u), nil
}

func checkSupport(codecs []string) bool {
	for _, codec := range codecs {
		if !strings.HasPrefix(codec, "avc1.") &&
			!strings.HasPrefix(codec, "hvc1.") &&
			!strings.HasPrefix(codec, "hev1.") &&
			!strings.HasPrefix(codec, "mp4a.") &&
			codec != "opus" {
			return false
		}
	}
	return true
}

func pickLeadingPlaylist(variants []*playlist.MultivariantVariant) *playlist.MultivariantVariant {
	var candidates []*playlist.MultivariantVariant //nolint:prealloc
	for _, v := range variants {
		if !checkSupport(v.Codecs) {
			continue
		}
		candidates = append(candidates, v)
	}
	if candidates == nil {
		return nil
	}

	// pick the variant with the greatest bandwidth
	var leadingPlaylist *playlist.MultivariantVariant
	for _, v := range candidates {
		if leadingPlaylist == nil ||
			v.Bandwidth > leadingPlaylist.Bandwidth {
			leadingPlaylist = v
		}
	}
	return leadingPlaylist
}
