package aria2go

import (
	"encoding/json"
	"log"
	"testing"
)

var (
	rpcSecret     = ""
	testClient, _ = NewClient("http://localhost:6800/jsonrpc", rpcSecret, nil)
)

type DummyNotifier struct{}

func (DummyNotifier) OnDownloadStart(events []Event)      { log.Printf("%s started.", events) }
func (DummyNotifier) OnDownloadPause(events []Event)      { log.Printf("%s paused.", events) }
func (DummyNotifier) OnDownloadStop(events []Event)       { log.Printf("%s stopped.", events) }
func (DummyNotifier) OnDownloadComplete(events []Event)   { log.Printf("%s completed.", events) }
func (DummyNotifier) OnDownloadError(events []Event)      { log.Printf("%s error.", events) }
func (DummyNotifier) OnBtDownloadComplete(events []Event) { log.Printf("bt %s completed.", events) }

func TestClient_AddURI(t *testing.T) {
	var (
		uris = []string{
			"https://github.com/prometheus/prometheus/releases/download/v2.53.0/prometheus-2.53.0.darwin-amd64.tar.gz",
		}
		options = map[string]any{
			"dir":                       "/root/downloads",
			"out":                       "prometheus-darwin.tar.gz",
			"http-proxy":                "http://127.0.0.1:7890",
			"https-proxy":               "http://127.0.0.1:7890",
			"split":                     5,
			"max-connection-per-server": 1,
			"header": []string{
				"user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
				"cookie: cookie_str_123",
			},
		}
	)

	gid, err := testClient.AddURI(uris, options, &DummyNotifier{})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("gid", gid)

	select {}
}

func TestClient_TellStatus(t *testing.T) {
	statusInfo, err := testClient.TellStatus("c28fd51f7b429579", "gid", "status")
	if err != nil {
		t.Log(err)
		return
	}

	t.Log(statusInfo.Gid, statusInfo.Status)
}

func TestClient_TellStopped(t *testing.T) {
	list, err := testClient.TellStopped(0, 100, "errorCode", "errorMessage", "gid", "status")
	if err != nil {
		t.Log(err)
		return
	}

	b, _ := json.Marshal(list)
	t.Logf("%s\n", b)
}

func TestClient_GetGlobalStat(t *testing.T) {
	globalStat, err := testClient.GetGlobalStat()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v\n", globalStat)
}

func TestClient_ListMethods(t *testing.T) {
	methods, err := testClient.ListMethods()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(methods)
	t.Log(len(methods))
}
