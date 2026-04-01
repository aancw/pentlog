package handlers

import (
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
)

func configManager() *config.ConfigManager {
	return config.Manager()
}

func formatSize(bytes int64) string {
	return utils.FormatBytes(bytes)
}
