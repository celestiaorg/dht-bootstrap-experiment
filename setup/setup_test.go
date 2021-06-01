package main

import (
	"testing"

	llconfig "github.com/lazyledger/lazyledger-core/config"
	"github.com/stretchr/testify/require"
)

const testConfigPath = "test-confing.toml"

func TestOpenAndSaveConfig(t *testing.T) {
	saveLazyConfig(testConfigPath, llconfig.DefaultConfig())
	cfg, err := loadLazyConfig(testConfigPath)
	cfg.LogFormat = "plain"
	cfg.TxIndex = &llconfig.TxIndexConfig{}
	saveLazyConfig(testConfigPath, cfg)
	require.NoError(t, err)

}
