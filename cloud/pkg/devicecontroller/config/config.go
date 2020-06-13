package config

import (
	"sync"

	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha2"
)

var Config Configure
var once sync.Once

type Configure struct {
	v1alpha2.DeviceController
	KubeAPIConfig v1alpha2.KubeAPIConfig
}

func InitConfigure(dc *v1alpha2.DeviceController, kubeAPIConfig *v1alpha2.KubeAPIConfig) {
	once.Do(func() {
		Config = Configure{
			DeviceController: *dc,
			KubeAPIConfig:    *kubeAPIConfig,
		}
	})
}
