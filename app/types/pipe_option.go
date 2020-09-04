package types

type PipeOption struct {
	Identity string `yaml:"identity"`
	Service  string `yaml:"service"`
	Validate string `yaml:"validate"`
	Topic    string `yaml:"topic"`
}
