package aria2go

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lysander66/zephyr/pkg/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

const (
	methodAddUri               = "aria2.addUri"
	methodAddTorrent           = "aria2.addTorrent"
	methodGetPeers             = "aria2.getPeers"
	methodAddMetalink          = "aria2.addMetalink"
	methodRemove               = "aria2.remove"
	methodPause                = "aria2.pause"
	methodForcePause           = "aria2.forcePause"
	methodPauseAll             = "aria2.pauseAll"
	methodForcePauseAll        = "aria2.forcePauseAll"
	methodUnpause              = "aria2.unpause"
	methodUnpauseAll           = "aria2.unpauseAll"
	methodForceRemove          = "aria2.forceRemove"
	methodChangePosition       = "aria2.changePosition"
	methodTellStatus           = "aria2.tellStatus"
	methodGetUris              = "aria2.getUris"
	methodGetFiles             = "aria2.getFiles"
	methodGetServers           = "aria2.getServers"
	methodTellActive           = "aria2.tellActive"
	methodTellWaiting          = "aria2.tellWaiting"
	methodTellStopped          = "aria2.tellStopped"
	methodGetOption            = "aria2.getOption"
	methodChangeUri            = "aria2.changeUri"
	methodChangeOption         = "aria2.changeOption"
	methodGetGlobalOption      = "aria2.getGlobalOption"
	methodChangeGlobalOption   = "aria2.changeGlobalOption"
	methodPurgeDownloadResult  = "aria2.purgeDownloadResult"
	methodRemoveDownloadResult = "aria2.removeDownloadResult"
	methodGetVersion           = "aria2.getVersion"
	methodGetSessionInfo       = "aria2.getSessionInfo"
	methodShutdown             = "aria2.shutdown"
	methodForceShutdown        = "aria2.forceShutdown"
	methodGetGlobalStat        = "aria2.getGlobalStat"
	methodSaveSession          = "aria2.saveSession"
	methodMultiCall            = "system.multicall"
	methodListMethods          = "system.listMethods"
	methodListNotifications    = "system.listNotifications"
)

type GlobalStat struct {
	DownloadSpeed   string `json:"downloadSpeed"`
	NumActive       string `json:"numActive"`
	NumStopped      string `json:"numStopped"`
	NumStoppedTotal string `json:"numStoppedTotal"`
	NumWaiting      string `json:"numWaiting"`
	UploadSpeed     string `json:"uploadSpeed"`
}

type Event struct {
	Gid string `json:"gid"`
}

type StatusInfo struct {
	CompletedLength string `json:"completedLength"`
	Connections     string `json:"connections"`
	Dir             string `json:"dir"`
	DownloadSpeed   string `json:"downloadSpeed"`
	ErrorCode       string `json:"errorCode"`
	ErrorMessage    string `json:"errorMessage"`
	Files           []struct {
		CompletedLength string `json:"completedLength"`
		Index           string `json:"index"`
		Length          string `json:"length"`
		Path            string `json:"path"`
		Selected        string `json:"selected"`
		Uris            []struct {
			Status string `json:"status"`
			Uri    string `json:"uri"`
		} `json:"uris"`
	} `json:"files"`
	Gid          string `json:"gid"`
	NumPieces    string `json:"numPieces"`
	PieceLength  string `json:"pieceLength"`
	Status       string `json:"status"`
	TotalLength  string `json:"totalLength"`
	UploadLength string `json:"uploadLength"`
	UploadSpeed  string `json:"uploadSpeed"`
}

// Notifier handles rpc notification from aria2 server
type Notifier interface {
	OnDownloadStart([]Event)
	OnDownloadPause([]Event)
	OnDownloadStop([]Event)
	OnDownloadComplete([]Event)
	OnDownloadError([]Event)
	OnBtDownloadComplete([]Event)
}

type Client struct {
	secret    string
	rpcClient *jsonrpc.Client
}

type Option func(o *Client)

func NewClient(endpoint, rpcSecret string, notifier Notifier) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{
		secret:    rpcSecret,
		rpcClient: jsonrpc.NewClient(endpoint),
	}

	if notifier != nil {
		u.Scheme = "ws"
		go c.setNotifier(u.String(), notifier)
	}
	return c, nil
}

func (c *Client) setNotifier(endpoint string, notifier Notifier) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		slog.Error("websocket.Dial", "err", err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("read:", "err", err)
				return
			}
			method := gjson.GetBytes(message, "method").String()
			if method != "" {
				var params []Event
				if err = json.Unmarshal([]byte(gjson.GetBytes(message, "params").String()), &params); err != nil {
					slog.Error("read:", "err", err, "message", message)
					return
				}
				switch method {
				case "aria2.onDownloadStart":
					notifier.OnDownloadStart(params)
				case "aria2.onDownloadPause":
					notifier.OnDownloadPause(params)
				case "aria2.onDownloadStop":
					notifier.OnDownloadStop(params)
				case "aria2.onDownloadComplete":
					notifier.OnDownloadComplete(params)
				case "aria2.onDownloadError":
					notifier.OnDownloadError(params)
				case "aria2.onBtDownloadComplete":
					notifier.OnBtDownloadComplete(params)
				}
			} else {
				slog.Info("recv: " + string(message))
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			slog.Info("interrupt")
			// Cleanly close the connection by sending a close message and then waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				slog.Error("write close:", "err", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (c *Client) token() string {
	return "token:" + c.secret
}

// AddURI
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.addUri
func (c *Client) AddURI(ctx context.Context, uris []string, options ...any) (string, error) {
	var params []any
	if c.secret != "" {
		params = append(params, c.token())
	}
	params = append(params, uris)
	if options != nil {
		params = append(params, options...)
	}

	req := jsonrpc.NewRequest(methodAddUri, params, time.Now().UnixNano())
	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetString()
}

// TellStatus | active waiting paused error complete removed
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.tellStatus
func (c *Client) TellStatus(ctx context.Context, gid string, keys ...string) (statusInfo StatusInfo, err error) {
	var params []any
	if c.secret != "" {
		params = append(params, c.token())
	}
	params = append(params, gid)
	if keys != nil {
		params = append(params, keys)
	}
	req := jsonrpc.NewRequest(methodTellStatus, params, time.Now().UnixNano())
	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		return
	}

	err = resp.GetAny(&statusInfo)
	return
}

func (c *Client) TellStopped(ctx context.Context, offset, num int, keys ...string) (list []StatusInfo, err error) {
	var params []any
	if c.secret != "" {
		params = append(params, c.token())
	}
	params = append(params, offset, num)
	if keys != nil {
		params = append(params, keys)
	}
	req := jsonrpc.NewRequest(methodTellStopped, params, time.Now().UnixNano())
	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		return
	}

	err = resp.GetAny(&list)
	return
}

func (c *Client) GetGlobalStat(ctx context.Context) (stat GlobalStat, err error) {
	var params []any
	if c.secret != "" {
		params = append(params, c.token())
	}
	req := jsonrpc.NewRequest(methodGetGlobalStat, params, time.Now().UnixNano())
	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		return
	}

	err = resp.GetAny(&stat)
	return
}

func (c *Client) ListMethods(ctx context.Context) (methods []string, err error) {
	req := jsonrpc.NewRequest(methodListMethods, nil, time.Now().UnixNano())
	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		return
	}

	err = resp.GetAny(&methods)
	return
}
