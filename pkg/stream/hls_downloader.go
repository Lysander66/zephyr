package stream

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Lysander66/zephyr/pkg/znet"
	"github.com/bluenviron/gohlslib/pkg/playlist"
)

type Progress struct {
	Downloaded int
	Total      int
}

type HLSDownloader struct {
	URI         string
	Name        string
	OutputDir   string
	TempDir     string
	NumParallel uint
	AutoCleanup bool
	ProgressCh  chan Progress
	Headers     map[string]string
	total       int
}

func NewHLSDownloader(uri, name, outputDir string, numParallel uint) *HLSDownloader {
	return &HLSDownloader{
		URI:         uri,
		Name:        name,
		OutputDir:   outputDir,
		TempDir:     filepath.Join(outputDir, "temp"),
		NumParallel: numParallel,
		AutoCleanup: true,
		ProgressCh:  make(chan Progress, 1000),
	}
}

func (d *HLSDownloader) Download() error {
	defer close(d.ProgressCh)

	err := os.MkdirAll(d.OutputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	err = os.MkdirAll(d.TempDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}

	now := time.Now()
	client := znet.New()
	client.SetHeaders(d.Headers)

	var segmentExt string
	var n atomic.Int32
	limiter := make(chan struct{}, d.NumParallel)

	p := &HLSPlayer{
		URI: d.URI,
		OnPlaylistDownloaded: func(b []byte, pl playlist.Playlist) {
			switch pls := pl.(type) {
			case *playlist.Multivariant:
				writeFile(filepath.Join(d.TempDir, "variant.m3u8"), b)
			case *playlist.Media:
				d.total = len(pls.Segments)
				// deep copy
				clonedPls := &playlist.Media{}
				data, _ := json.Marshal(pls)
				json.Unmarshal(data, clonedPls)
				for i, seg := range clonedPls.Segments {
					if segmentExt == "" {
						u, _ := url.Parse(seg.URI)
						segmentExt = path.Ext(u.Path)
						slog.Debug("file extension ", "ext", segmentExt)
					}
					seg.URI = fmt.Sprintf("%d.%s", i+1, segmentExt)
				}
				data, _ = clonedPls.Marshal()
				writeFile(filepath.Join(d.TempDir, "playlist.m3u8"), data)
				writeFile(filepath.Join(d.TempDir, "index.m3u8"), b)
			}
		},
		FetchSegment: func(url string, i int) error {
			limiter <- struct{}{}
			go func() {
				defer func() { <-limiter }()

				filename := filepath.Join(d.TempDir, fmt.Sprintf("%d.%s", i+1, segmentExt))
				if err := downloadSegment(client, url, filename); err != nil {
					slog.Error("downloadSegment", "err", err, "url", url)
					return
				}
				n.Add(1)
				d.reportProgress(int(n.Load()), d.total)
			}()
			return nil
		},
	}

	err = p.Run()
	if err != nil {
		return err
	}

	err = p.Wait()
	if err != nil {
		return err
	}
	slog.Info("Download completed", "elapsed", time.Since(now).String())

	filename := d.Name + ".mp4"
	if err := d.mergeToMP4(filename); err != nil {
		return err
	}
	slog.Info("Merge to "+filepath.Join(d.OutputDir, filename), "elapsed", time.Since(now).String())

	if d.AutoCleanup {
		return d.Cleanup()
	}
	return nil
}

func (d *HLSDownloader) mergeToMP4(filename string) error {
	args := fmt.Sprintf("cd %s && ffmpeg -i playlist.m3u8 -c copy %s && mv %[2]s %s", d.TempDir, filename, d.OutputDir)
	if err := exec.Command("/bin/sh", "-c", args).Run(); err != nil {
		slog.Info(args)
		slog.Error("mergeToMP4", "err", err)
		return err
	}
	return nil
}

func (d *HLSDownloader) Cleanup() error {
	return os.RemoveAll(d.TempDir)
}

func (d *HLSDownloader) reportProgress(downloaded, total int) {
	slog.Debug("reportProgress", "-->", fmt.Sprintf("%d/%d", downloaded, total))
	d.ProgressCh <- Progress{Downloaded: downloaded, Total: total}
}

func downloadSegment(client *znet.Client, url, filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		//nop
	} else {
		slog.Debug("File exists, skip downloading", "filename", filename)
		return nil
	}

	resp, err := client.R().GetWithRetries(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("status: %d", resp.StatusCode())
	}

	return os.WriteFile(filename, resp.Body(), 0644)
}

func writeFile(name string, data []byte) {
	err := os.WriteFile(name, data, 0644)
	if err != nil {
		slog.Error("Failed to write file", "err", err)
	}
}

func extractFileNameFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// return path.Base(u.Path), nil
	return strings.Replace(u.Path, "/", "_", -1), nil
}
