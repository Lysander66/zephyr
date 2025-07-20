package streamurl

import (
	"fmt"
	"testing"
	"time"
)

func testProvider(t *testing.T, provider string) {
	agent, err := NewAgent(
		provider,
		WithPushConfig("push.example.com", "your_push_key"),
		WithPullConfig("pull.example.com", "your_pull_key"),
		WithHTTPS(true),
	)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	appName := "live"
	streamName := "stream123"
	expireAt := time.Now().Unix() + 3600

	fmt.Printf("--- Testing Provider: %s ---\n", provider)

	pushURL := agent.PublishingAddress(appName, streamName, expireAt)
	fmt.Printf("Push URL: %s\n", pushURL)
	if pushURL == "" {
		t.Errorf("PublishingAddress should not be empty")
	}

	rtmpURL := agent.RTMPPlayUrl(appName, streamName, expireAt)
	fmt.Printf("RTMP URL: %s\n", rtmpURL)
	if rtmpURL == "" {
		t.Errorf("RTMPPlayUrl should not be empty")
	}

	flvURL := agent.FlvPlayUrl(appName, streamName, expireAt)
	fmt.Printf("FLV URL: %s\n", flvURL)
	if flvURL == "" {
		t.Errorf("FlvPlayUrl should not be empty")
	}

	hlsURL := agent.HlsPlayUrl(appName, streamName, expireAt)
	fmt.Printf("HLS URL: %s\n", hlsURL)
	if hlsURL == "" {
		t.Errorf("HlsPlayUrl should not be empty")
	}

	fmt.Println("---------------------------------")
}

func TestAgent(t *testing.T) {
	providers := []string{"aliyun", "tencent", "wangsu", "huawei"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			testProvider(t, provider)
		})
	}
}
