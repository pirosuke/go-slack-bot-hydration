package database

type (
	// DbConfig describes database connection config.
	DbConfig struct {
		Client     string `json:"client"`
		Connection struct {
			Host     string `json:"host"`
			Port     int64  `json:"port"`
			Database string `json:"database"`
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"connection"`
	}
)
