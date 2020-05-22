package config

import "testing"

func TestConfigInit(t *testing.T) {
	if Config.APIName == "" {
		t.Error("Config.APIName is empty")
		return
	}
	// t.Logf("Config: %v", Config)
}
