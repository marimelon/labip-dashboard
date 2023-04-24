package main

import (
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	client := NewNeburaCliend(
		os.Getenv("NEBURA_USER"),
		os.Getenv("NEBURA_PASSWORD"),
		os.Getenv("NEBURA_ENDPOINT"),
	)
	vmmac, err := client.GetVMNameWithMac()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v", vmmac)
}
