package common

type Inject struct {
	Values *Values
}

type Values struct {
	Address string `yaml:"address"`
}
