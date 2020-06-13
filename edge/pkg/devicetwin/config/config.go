package config

import (
	"sync"

	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha2"
)

var Config Configure
var once sync.Once

type Configure struct {
	v1alpha2.DeviceTwin
	NodeName string
}

func InitConfigure(deviceTwin *v1alpha2.DeviceTwin, nodeName string) {
	once.Do(func() {
		Config = Configure{
			DeviceTwin: *deviceTwin,
			NodeName:   nodeName,
		}
	})
}

func Get() *Configure {
	return &Config
}
