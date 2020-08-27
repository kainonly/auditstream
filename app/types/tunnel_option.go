package types

type TunnelOption struct {
	SrcIp   string `yaml:"src_ip"`
	SrcPort uint32 `yaml:"src_port"`
	DstIp   string `yaml:"dst_ip"`
	DstPort uint32 `yaml:"dst_port"`
}
