package common

import "errors"

var (
	CertsFromPEMError = errors.New("根证书加载失败")
)
