package options

type PipeOption struct {
	Identity string `yaml:"identity"`
	Index    string `yaml:"index"`
	Validate string `yaml:"validate"`
	Topic    string `yaml:"topic"`
	Key      string `yaml:"key"`
}
