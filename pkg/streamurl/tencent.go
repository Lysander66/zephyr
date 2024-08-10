package streamurl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lysander66/zephyr/pkg/zcrypto"
)

// TxGenerator 腾讯
// https://cloud.tencent.com/document/product/267/35257
type TxGenerator struct{}

func (t TxGenerator) PublishingAddress(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s", app, stream) + t.secret(key, stream, exp)
}

func (t TxGenerator) RTMPPlayUrl(key, app, stream string, exp int64) string {
	return "Not implemented"
}

func (t TxGenerator) FlvPlayUrl(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s.%s", app, stream, "flv") + t.secret(key, stream, exp)
}

func (t TxGenerator) HlsPlayUrl(key, app, stream string, exp int64) string {
	return fmt.Sprintf("/%s/%s.%s", app, stream, "m3u8") + t.secret(key, stream, exp)
}

func (t TxGenerator) secret(key, stream string, exp int64) string {
	if key != "" && exp > 0 {
		txTime := strings.ToUpper(strconv.FormatInt(exp, 16))
		txSecret := zcrypto.MD5Sum(key + stream + txTime)
		return "?txSecret=" + txSecret + "&txTime=" + txTime
	}
	return ""
}
