package config

import "go.thethings.network/lorawan-stack/v3/pkg/log"

func (c *Config) LogFields() log.Fielder {
	return log.Fields(
		"export_vars", c.ExportVars,
		"export_session", c.ExportSession,
		"insecure", c.insecure,
		"url", c.url,
	)
}
