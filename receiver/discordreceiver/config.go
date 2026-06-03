package discordreceiver

type Config struct {
	GuildIDs   []string `mapstructure:"guild_ids"`
	ChannelIDs []string `mapstructure:"channel_ids"`
}

func (cfg *Config) Validate() error {
	return nil
}
