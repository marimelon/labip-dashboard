package main

import (
	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

type NeburaCliend struct{ *goca.Client }

func NewNeburaCliend(user, password, endpoint string) (c *NeburaCliend) {
	client := goca.NewDefaultClient(
		goca.NewConfig(user, password, "http://sddlab-cloud-01.is.oit.ac.jp:2633/RPC2"),
	)
	return &NeburaCliend{Client: client}
}

type VMMacResult struct {
	Type string
	ID   int
	Name string
	Mac  string
}

func (c *NeburaCliend) GetVMNameWithMac() ([]VMMacResult, error) {
	controller := goca.NewController(c.Client)
	pool, err := controller.VMs().Info()
	if err != nil {
		return nil, err
	}

	res := []VMMacResult{}

	for _, vm := range pool.VMs {
		for _, nic := range vm.Template.GetNICs() {
			mac, err := nic.GetStr("MAC")
			if err != nil {
				return nil, err
			}
			res = append(res, VMMacResult{"Nebura", vm.ID, vm.Name, mac})
		}
	}

	return res, nil
}
