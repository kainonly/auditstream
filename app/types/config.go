package types

type Config struct {
	Debug  bool   `yaml:"debug"`
	Listen string `yaml:"listen"`
}
