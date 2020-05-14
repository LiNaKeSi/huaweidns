package huaweidns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HuaweiDNSClient struct {
	s        Signer
	endpoint string
	zoneId   string
}

type RecordInfo struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Records     []string `json:"records"`

	// ZoneId string `json:"zone_id"`
	// ZoneName string `json:"zone_name"`
}

func (c *HuaweiDNSClient) AddDomainRecord(recordName string, Type string, value string) error {
	if Type == "TXT" && value[0] != '"' {
		value = fmt.Sprintf("%q", value)
	}
	var ret interface{}
	err := c.post(fmt.Sprintf("/v2/zones/%s/recordsets", c.zoneId),
		RecordInfo{
			Name:    recordName,
			Type:    Type,
			Records: []string{value},
		},
		&ret,
	)
	return err
}

func (c *HuaweiDNSClient) DeleteDomainRecord(recordName string, Type string) error {
	rs, err := c.List(recordName)
	if err != nil {
		return err
	}
	for _, r := range rs {
		var resp interface{}
		err = c._delete(fmt.Sprintf("/v2/zones/%s/recordsets/%s", c.zoneId, r.Id), &resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *HuaweiDNSClient) List(name string) ([]RecordInfo, error) {
	var resp struct {
		RecordSets []RecordInfo `json:"recordsets"`
	}
	err := c.get(fmt.Sprintf("/v2/zones/%s/recordsets?name=%s", c.zoneId, name), &resp)
	return resp.RecordSets, err
}

func NewHuaweiDNSClient(appKey, appSecret string, domain string) (*HuaweiDNSClient, error) {
	s := Signer{
		Key:    appKey,
		Secret: appSecret,
	}
	c := &HuaweiDNSClient{
		s:        s,
		endpoint: "dns.myhuaweicloud.com",
	}
	return c, c.initZoneId(domain)
}

func (c *HuaweiDNSClient) initZoneId(domain string) error {
	var resp struct {
		Zones []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"zones"`
	}
	err := c.get(fmt.Sprintf("/v2/zones?name=%s&limit=1", domain), &resp)
	if err != nil {
		return err
	}
	if len(resp.Zones) == 0 {
		return fmt.Errorf("Can't found zoneId for %q %q", domain, resp.Zones)
	}
	c.zoneId = resp.Zones[0].Id
	return nil
}

func (c *HuaweiDNSClient) get(resourcePath string, value interface{}) error {
	r, err := http.NewRequest("GET",
		fmt.Sprintf("https://%s%s", c.endpoint, resourcePath),
		nil,
	)
	if err != nil {
		return err
	}
	return c.do(r, value)
}

func (c *HuaweiDNSClient) _delete(resourcePath string, value interface{}) error {
	r, err := http.NewRequest("DELETE",
		fmt.Sprintf("https://%s%s", c.endpoint, resourcePath),
		nil,
	)
	if err != nil {
		return err
	}
	return c.do(r, value)
}

func (c *HuaweiDNSClient) post(resourcePath string, body interface{}, value interface{}) error {
	b := bytes.NewBuffer(nil)
	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		return err
	}
	r, err := http.NewRequest("POST",
		fmt.Sprintf("https://%s%s", c.endpoint, resourcePath),
		b,
	)
	if err != nil {
		return err
	}
	return c.do(r, value)
}

func (c *HuaweiDNSClient) do(r *http.Request, value interface{}) error {
	r.Header.Add("content-type", "application/json")

	c.s.Sign(r)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("OpenURL %v failed %q %q", r, resp.Status, string(bs))
	}
	return json.NewDecoder(resp.Body).Decode(value)
}
