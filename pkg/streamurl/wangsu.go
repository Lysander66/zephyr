package streamurl

import (
	"fmt"
	"strconv"

	"github.com/Lysander66/zephyr/pkg/zcrypto"
)

// WsGenerator 网宿
// https://www.wangsu.com/document/livestream/171
type WsGenerator struct{}

func (w WsGenerator) PublishingAddress(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s", app, stream) + w.secret(key, app, stream, exp)
}

func (w WsGenerator) RTMPPlayUrl(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s", app, stream) + w.secret(key, app, stream, exp)
}

func (w WsGenerator) FlvPlayUrl(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s.%s", app, stream, "flv") + w.secret(key, app, stream, exp)
}

func (w WsGenerator) HlsPlayUrl(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s.%s", app, stream, "m3u8") + w.secret(key, app, stream, exp)
}

func (w WsGenerator) secret(key, app, stream string, exp int64) string {
	if key != "" && exp > 0 {
		wsTime := strconv.FormatInt(exp, 16)
		wsSecret := zcrypto.MD5Sum(fmt.Sprintf("/%s/%s%s%s", app, stream, key, wsTime))
		return "?wsSecret=" + wsSecret + "&wsABSTime=" + wsTime
	}
	return ""
}
