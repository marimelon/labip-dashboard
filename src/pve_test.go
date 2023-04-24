package main

import (
	"fmt"
	"os"
	"testing"
)

func SetupPVE() *PVEClient {
	pve := NewPVECliend(
		os.Getenv("NEBURA_USER"),
		os.Getenv("PVE_API_USER"),
		os.Getenv("PVE_API_KEY"),
	)
	return pve
}

func TestPVE(t *testing.T) {
	pve := SetupPVE()
	nodes, err := pve.GetAllNodeIds()
	if err != nil {
		t.Error(err)
	}
	fmt.Print("&v", nodes)
}

func TestPVE2(t *testing.T) {
	pve := SetupPVE()

	macs, err := pve.GetVMNameWithMac()
	if err != nil {
		t.Error(err)
	}
	fmt.Print("&v", macs)
}
