package types

type PipeOption struct {
	Identity string `yaml:"identity"`
	Validate string `yaml:"validate"`
	Topic    string `yaml:"topic"`
	Key      string `yaml:"key"`
}
