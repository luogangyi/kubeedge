package config

import (
	"sync"

	configv1alpha2 "github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha2"
)

var Config Configure
var once sync.Once

type Configure struct {
	KubeAPIConfig  *configv1alpha2.KubeAPIConfig
	SyncController *configv1alpha2.SyncController
}

func InitConfigure(sc *configv1alpha2.SyncController, kubeAPIConfig *configv1alpha2.KubeAPIConfig) {
	once.Do(func() {
		Config = Configure{
			KubeAPIConfig:  kubeAPIConfig,
			SyncController: sc,
		}
	})
}
