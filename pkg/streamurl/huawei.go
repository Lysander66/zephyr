package streamurl

import (
	"fmt"
	"strconv"

	"github.com/Lysander66/zephyr/pkg/zcrypto"
)

// HwGenerator 华为
// https://support.huaweicloud.com/iLive-live/live010007.html
// https://support.huaweicloud.com/iLive-live/live_01_0049.html#section7
type HwGenerator struct{}

func (g *HwGenerator) PublishingAddress(key, app, stream string, exp int64) string {
	return g.generateURL("rtmp", key, app, stream, exp, true)
}

func (g *HwGenerator) RTMPPlayUrl(key, app, stream string, exp int64) string {
	return g.generateURL("rtmp", key, app, stream, exp, false)
}

func (g *HwGenerator) FlvPlayUrl(key, app, stream string, exp int64) string {
	return g.generateURL("flv", key, app, stream, exp, false)
}

func (g *HwGenerator) HlsPlayUrl(key, app, stream string, exp int64) string {
	return g.generateURL("m3u8", key, app, stream, exp, false)
}

func (g *HwGenerator) generateURL(proto, key, app, stream string, exp int64, isPush bool) string {
	hwTime := strconv.FormatInt(exp, 16)
	hwSecret := zcrypto.HmacSHA256(stream+hwTime, key)

	path := fmt.Sprintf("/%s/%s", app, stream)
	if !isPush && proto != "rtmp" {
		path = fmt.Sprintf("%s.%s", path, proto)
	}

	return fmt.Sprintf("%s?hwSecret=%s&hwTime=%s", path, hwSecret, hwTime)
}
