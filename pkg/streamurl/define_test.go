package streamurl

import (
	"testing"
	"time"
)

func Test_PublishingAddress(t *testing.T) {
	var (
		name        = "aliyun"
		pushDomain  = "sportpush.sportlive.vip"
		pushAuthKey = "RAmE8kMNx1BpbAea"
		pullDomain  = "sport.esptv666.com"
		pullAuthKey = "5wKMXZPiXSEWrh7HMQeX"
		appName     = "live"
		streamName  = "szPFCyay91"
		expireAt    = time.Now().Add(24 * time.Hour).Unix()
	)
	agent, err := NewAgent(name, pushDomain, pushAuthKey, pullDomain, pullAuthKey, true)
	if err != nil {
		t.Error(err)
		return
	}

	addr := agent.PublishingAddress(appName, streamName, expireAt)
	t.Log(addr)

	rtmpPlayUrl := agent.RTMPPlayUrl(appName, streamName, expireAt)
	t.Log(rtmpPlayUrl)

	flvPlayUrl := agent.FlvPlayUrl(appName, streamName, expireAt)
	t.Log(flvPlayUrl)

	hlsPlayUrl := agent.HlsPlayUrl(appName, streamName, expireAt)
	t.Log(hlsPlayUrl)
}
