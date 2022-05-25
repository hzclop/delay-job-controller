package config

import (
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type Env string

const (
	dev Env = "dev"
)

type Config struct {
	Env       string `properties:"env,default=dev"`
	PprofPort string `properties:"pprof_port,default=:7424"`
}

var (
	cfg  *Config
	kube *rest.Config
)

func GetCfg() *Config {
	return cfg
}

func GetKubeCfg() *rest.Config {
	return kube
}

func Init() error {
	var configure string
	var kubeConfig string
	flag.StringVar(&configure, "configure", "D:\\Go\\src\\github.com\\hhzhhzhhz\\delay-job-controller\\etc\\config.properties", "configure file for k8s-job-scheduler")
	flag.StringVar(&kubeConfig, "kubeconfig", "", "configure file for kubernetes ApiServer")
	flag.Parse()
	p := properties.MustLoadFile(configure, properties.UTF8)
	config := &Config{}
	if err := p.Decode(config); err != nil {
		return fmt.Errorf("Server.Init Config.Decode failed cause=%s", err.Error())
	}
	cfg = config
	var kluge *rest.Config
	var err error
	if kubeConfig == "" {
		kluge, err = rest.InClusterConfig()
		if err != nil {
			kubeConfig = filepath.Join(os.Getenv("HOME"), "/etc", "config")
			kluge, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
			if err != nil {
				return fmt.Errorf("Server.Init BuildConfigFromFlags kluge failed cause=%s", err.Error())
			}
		}
	} else {
		kluge, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return fmt.Errorf("Server.Init BuildConfigFromFlags kluge failed cause=%s", err.Error())
		}
	}
	kube = kluge
	return err
}
