package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type PVEClient struct {
	User     string
	Token    string
	Endpoint string
}

func NewPVECliend(endpoint, user, token string) *PVEClient {
	return &PVEClient{User: user, Endpoint: endpoint, Token: token}
}

func (c *PVEClient) GetApi(path string, respBody interface{}) (int, error) {
	req, err := http.NewRequest("GET", c.Endpoint+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "PVEAPIToken="+c.User+"="+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf(resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

type NodesResponse struct {
	ID string `json:"node"`
}

func (c *PVEClient) GetAllNodeIds() ([]NodesResponse, error) {
	type NodesResponseData struct {
		Data []NodesResponse `json:"data"`
	}

	var response *NodesResponseData
	_, err := c.GetApi("/nodes", &response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

type VmsResponse struct {
	VmId int    `json:"vmid"`
	Name string `json:"name"`
}

func (c *PVEClient) GetVms(node string) ([]VmsResponse, error) {
	type VmsResponseData struct {
		Data []VmsResponse `json:"data"`
	}

	var response *VmsResponseData
	_, err := c.GetApi("/nodes/"+node+"/qemu", &response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

type VmConfigResponse struct {
	Net0 string `json:"net0"`
	Name string `json:"name"`
}

func (c *PVEClient) GetVMConfig(node string, vmid int) (*VmConfigResponse, error) {
	type VmConfigResponseData struct {
		Data VmConfigResponse `json:"data"`
	}

	var response *VmConfigResponseData
	_, err := c.GetApi("/nodes/"+node+"/qemu/"+fmt.Sprint(vmid)+"/config", &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

var macExp = regexp.MustCompile(`virtio=(?P<mac>[a-zA-Z0-9:]+)`)

func (c *PVEClient) GetVMNameWithMac() ([]VMMacResult, error) {
	nodes, err := c.GetAllNodeIds()
	if err != nil {
		return nil, err
	}

	res := []VMMacResult{}
	for _, node := range nodes {
		vms, err := c.GetVms(node.ID)
		if err != nil {
			return nil, err
		}
		for _, vm := range vms {
			conf, err := c.GetVMConfig(node.ID, vm.VmId)
			if err != nil {
				return nil, err
			}
			mac := macExp.FindStringSubmatch(conf.Net0)
			if len(mac) < 2 {
				return nil, fmt.Errorf("not found mac address")
			}

			res = append(res, VMMacResult{"PVE", vm.VmId, vm.Name, mac[1]})
		}
	}

	return res, nil
}
