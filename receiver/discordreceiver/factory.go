package discordreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var typeStr = component.MustNewType("discord")

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createLogsReceiver(_ context.Context, settings receiver.Settings, baseCfg component.Config, consumer consumer.Logs) (receiver.Logs, error) {
	return newReceiver(baseCfg.(*Config), settings, consumer), nil
}
