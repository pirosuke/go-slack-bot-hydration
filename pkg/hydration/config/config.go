package config

import (
	"github.com/pirosuke/slack-bot-hydration/internal/database"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/slack"
)

type (
	// Config describes global server config.
	Config struct {
		Db         database.DbConfig `json:"db"`
		LogDirPath string            `json:"log_dir"`
		Slack      slack.Config      `json:"slack"`
	}
)
