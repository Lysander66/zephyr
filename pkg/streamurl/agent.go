package streamurl

import (
	"fmt"
	"strings"
)

// Generator StreamURLGenerator
type Generator interface {
	PublishingAddress(key, app, stream string, exp int64) string
	RTMPPlayUrl(key, app, stream string, exp int64) string
	FlvPlayUrl(key, app, stream string, exp int64) string
	HlsPlayUrl(key, app, stream string, exp int64) string
}

type Agent struct {
	Name      string
	generator Generator
	config    *Config
}

// Config 配置选项
type Config struct {
	PushDomain  string
	PushAuthKey string
	PullDomain  string
	PullAuthKey string
	HTTPS       bool
}

// Option 配置函数类型
type Option func(*Config)

// WithPushConfig 设置推流配置
func WithPushConfig(domain, authKey string) Option {
	return func(c *Config) {
		c.PushDomain = strings.TrimSpace(domain)
		c.PushAuthKey = strings.TrimSpace(authKey)
	}
}

// WithPullConfig 设置拉流配置
func WithPullConfig(domain, authKey string) Option {
	return func(c *Config) {
		c.PullDomain = strings.TrimSpace(domain)
		c.PullAuthKey = strings.TrimSpace(authKey)
	}
}

// WithHTTPS 设置是否使用 HTTPS
func WithHTTPS(https bool) Option {
	return func(c *Config) {
		c.HTTPS = https
	}
}

// NewAgent 创建新的 Agent 实例
func NewAgent(name string, opts ...Option) (*Agent, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	agent := &Agent{
		Name:   strings.TrimSpace(name),
		config: config,
	}

	switch agent.Name {
	case "aliyun":
		agent.generator = &AliGenerator{}
	case "tencent":
		agent.generator = &TxGenerator{}
	case "wangsu":
		agent.generator = &WsGenerator{}
	case "huawei":
		agent.generator = &HwGenerator{}
	default:
		return nil, fmt.Errorf("%s is not supported yet", name)
	}
	return agent, nil
}

// PublishingPath 返回推流路径（不带域名）
func (agt *Agent) PublishingPath(appName, streamName string, expireAt int64) string {
	return agt.generator.PublishingAddress(agt.config.PushAuthKey, appName, streamName, expireAt)
}

// RTMPPlayPath 返回 RTMP 播放路径（不带域名）
func (agt *Agent) RTMPPlayPath(appName, streamName string, expireAt int64) string {
	return agt.generator.RTMPPlayUrl(agt.config.PullAuthKey, appName, streamName, expireAt)
}

// FlvPlayPath 返回 FLV 播放路径（不带域名）
func (agt *Agent) FlvPlayPath(appName, streamName string, expireAt int64) string {
	return agt.generator.FlvPlayUrl(agt.config.PullAuthKey, appName, streamName, expireAt)
}

// HlsPlayPath 返回 HLS 播放路径（不带域名）
func (agt *Agent) HlsPlayPath(appName, streamName string, expireAt int64) string {
	return agt.generator.HlsPlayUrl(agt.config.PullAuthKey, appName, streamName, expireAt)
}

// 以下是完整 URL 的方法，仅在配置了域名时可用

// PublishingAddress 返回完整推流地址
func (agt *Agent) PublishingAddress(appName, streamName string, expireAt int64) string {
	return "rtmp://" + agt.config.PushDomain + agt.PublishingPath(appName, streamName, expireAt)
}

// RTMPPlayUrl 返回完整 RTMP 播放地址
func (agt *Agent) RTMPPlayUrl(appName, streamName string, expireAt int64) string {
	return "rtmp://" + agt.config.PullDomain + agt.RTMPPlayPath(appName, streamName, expireAt)
}

// FlvPlayUrl 返回完整 FLV 播放地址
func (agt *Agent) FlvPlayUrl(appName, streamName string, expireAt int64) string {
	path := agt.FlvPlayPath(appName, streamName, expireAt)
	if agt.config.HTTPS {
		return "https://" + agt.config.PullDomain + path
	}
	return "http://" + agt.config.PullDomain + path
}

// HlsPlayUrl 返回完整 HLS 播放地址
func (agt *Agent) HlsPlayUrl(appName, streamName string, expireAt int64) string {
	path := agt.HlsPlayPath(appName, streamName, expireAt)
	if agt.config.HTTPS {
		return "https://" + agt.config.PullDomain + path
	}
	return "http://" + agt.config.PullDomain + path
}
