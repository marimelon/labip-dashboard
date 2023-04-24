package main

import (
	"bytes"
	"fmt"
	ouidb "labip-dashboard/go-ouitools"
	"sort"
	"strings"
	"time"
)

type IpInfo struct {
	Ip      string
	Mac     string
	Company string
}

func (h *Handler) get_ip_info() ([]IpInfo, error) {
	foundIPs, err := scan(h.cfg.InterfaceName, 5*time.Second)
	if err != nil {
		return nil, err
	}

	sort.Slice(foundIPs, func(i, j int) bool {
		return bytes.Compare(foundIPs[i].Ip, foundIPs[j].Ip) < 0
	})

	vmsmac := []VMMacResult{}
	if h.pve != nil {
		_vmsmac, err := h.pve.GetVMNameWithMac()
		if err != nil {
			return nil, err
		}
		vmsmac = append(vmsmac, _vmsmac...)
	}

	if h.nebura != nil {
		_vmsmac, err := h.nebura.GetVMNameWithMac()
		if err != nil {
			return nil, err
		}
		vmsmac = append(vmsmac, _vmsmac...)
	}

	db := ouidb.New("oui.txt")
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ipInfoList := []IpInfo{}
	for _, ip := range foundIPs {
		ipinfo := IpInfo{ip.Ip.String(), "", ""}
		if ip.Mac == nil {
			ipinfo.Mac = "★未使用"
			ipInfoList = append(ipInfoList, ipinfo)
			continue
		} else {
			ipinfo.Mac = ip.Mac.String()
		}

		// 仮想マシンから検索
		for _, vmmac := range vmsmac {
			if strings.EqualFold(ipinfo.Mac, vmmac.Mac) {
				ipinfo.Company = fmt.Sprintf("%s[%d]: %s", vmmac.Type, vmmac.ID, vmmac.Name)
			}
		}

		if ipinfo.Company == "" {
			// Vendorを検索
			name, err := db.VendorLookup(ip.Mac.String())
			if err == nil {
				fmt.Printf("%v", name)
				ipinfo.Company = name
			}
		}

		if ipinfo.Company == "" {
			ipinfo.Company = "不明"
		}

		ipInfoList = append(ipInfoList, ipinfo)
	}

	return ipInfoList, nil
}
