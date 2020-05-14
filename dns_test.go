package huaweidns

import (
	"encoding/json"
	"os"
	"testing"
)

type TestConfig struct {
	AppKey    string
	AppSecret string
}

const appDomain = "linakesi.com"

var cfg TestConfig

func init() {
	f, err := os.Open("secret.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		panic(err)
	}
}

func TestPing(t *testing.T) {
	c, err := NewHuaweiDNSClient(cfg.AppKey, cfg.AppSecret, appDomain)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := c.List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) == 0 {
		t.Fatal("Can't find any dns record")
	}
}

func TestAddDel(t *testing.T) {
	c, err := NewHuaweiDNSClient(cfg.AppKey, cfg.AppSecret, appDomain)
	if err != nil {
		t.Fatal(err)
	}
	tname := "testsnyh.linakesi.com"
	err = c.AddDomainRecord(tname, "TXT", "123")
	if err != nil {
		t.Fatal(err)
	}
	c.DeleteDomainRecord(tname, "TXT")
}
