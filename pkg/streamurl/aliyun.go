package streamurl

import (
	"fmt"

	"github.com/Lysander66/zephyr/pkg/zcrypto"
)

// AliGenerator 阿里
// https://help.aliyun.com/zh/live/developer-reference/ingest-and-streaming-urls
type AliGenerator struct{}

func (a AliGenerator) PublishingAddress(key, app, stream string, exp int64) string {
	if key == "" {
		return fmt.Sprintf("/%s/%s", app, stream)
	}
	str := fmt.Sprintf("/%s/%s-%d-0-0-%s", app, stream, exp, key)
	return fmt.Sprintf("/%s/%s?auth_key=%d-0-0-%s", app, stream, exp, zcrypto.MD5Sum(str))
}

func (a AliGenerator) RTMPPlayUrl(key, app, stream string, exp int64) string {
	if key == "" {
		return fmt.Sprintf("/%s/%s", app, stream)
	}
	str := fmt.Sprintf("/%s/%s-%d-0-0-%s", app, stream, exp, key)
	return fmt.Sprintf("/%s/%s?auth_key=%d-0-0-%s", app, stream, exp, zcrypto.MD5Sum(str))
}

func (a AliGenerator) FlvPlayUrl(key, app, stream string, exp int64) string {
	return a.playUrl(key, app, stream, "flv", exp)
}

func (a AliGenerator) HlsPlayUrl(key, app, stream string, exp int64) string {
	return a.playUrl(key, app, stream, "m3u8", exp)
}

func (a AliGenerator) playUrl(key, app, stream, format string, exp int64) string {
	if key == "" {
		return fmt.Sprintf("/%s/%s.%s", app, stream, format)
	}
	str := fmt.Sprintf("/%s/%s.%s-%d-0-0-%s", app, stream, format, exp, key)
	return fmt.Sprintf("/%s/%s.%s?auth_key=%d-0-0-%s", app, stream, format, exp, zcrypto.MD5Sum(str))
}
