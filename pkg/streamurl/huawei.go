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
	path := fmt.Sprintf("/%s/%s", app, stream)
	if !isPush && proto != "rtmp" {
		path = fmt.Sprintf("%s.%s", path, proto)
	}

	if key != "" && exp > 0 {
		hwTime := strconv.FormatInt(exp, 16)
		stringToSign := stream + hwTime
		hwSecret := zcrypto.HmacSHA256(stringToSign, key)
		return fmt.Sprintf("%s?hwSecret=%s&hwTime=%s", path, hwSecret, hwTime)
	}

	return path
}
