package common

import (
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	bufpool *sync.Pool
)

type (
	AppOption struct {
		Debug  bool   `yaml:"debug"`
		Listen string `yaml:"listen"`
		Pool   uint32 `yaml:"pool"`
	}
	ConnectOption struct {
		Host       string
		Port       uint32
		Username   string
		Password   string
		Key        []byte
		PassPhrase []byte
	}
	TunnelOption struct {
		SrcIp   string `yaml:"src_ip"`
		SrcPort uint32 `yaml:"src_port"`
		DstIp   string `yaml:"dst_ip"`
		DstPort uint32 `yaml:"dst_port"`
	}
	ConfigOption struct {
		Identity   string         `yaml:"identity"`
		Host       string         `yaml:"host"`
		Port       uint32         `yaml:"port"`
		Username   string         `yaml:"username"`
		Password   string         `yaml:"password"`
		Key        string         `yaml:"key"`
		PassPhrase string         `yaml:"pass_phrase"`
		Tunnels    []TunnelOption `yaml:"tunnels"`
	}
)

// Init buffer pool
func InitBufPool(size uint32) {
	bufpool = &sync.Pool{}
	bufpool.New = func() interface{} {
		return make([]byte, size*1024)
	}
}

// Use buffer pool io copy
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	buf := bufpool.Get().([]byte)
	defer bufpool.Put(buf)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

// Get Addr
func GetAddr(ip string, port uint) string {
	return ip + ":" + strconv.Itoa(int(port))
}

func autoload(identity string) string {
	return "./config/autoload/" + identity + ".yml"
}

func ListConfig() (list []ConfigOption, err error) {
	var files []os.FileInfo
	files, err = ioutil.ReadDir("./config/autoload")
	if err != nil {
		return
	}
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".yml" {
			var in []byte
			in, err = ioutil.ReadFile("./config/autoload/" + file.Name())
			if err != nil {
				return
			}
			var config ConfigOption
			err = yaml.Unmarshal(in, &config)
			if err != nil {
				return
			}
			list = append(list, config)
		}
	}
	return
}

func SaveConfig(config ConfigOption) (err error) {
	var out []byte
	out, err = yaml.Marshal(config)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(
		autoload(config.Identity),
		out,
		0644,
	)
	if err != nil {
		return
	}
	return
}

func RemoveConfig(identity string) error {
	return os.Remove(autoload(identity))
}
