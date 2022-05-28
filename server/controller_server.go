package server

import (
	"context"
	"fmt"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/delay-job-controller/config"
	"k8s.io/delay-job-controller/controller/delayjob"
	"k8s.io/delay-job-controller/infrastructure/kubernetes"
	"k8s.io/delay-job-controller/log"
	clientset "k8s.io/delay-job-controller/pkg/generated/clientset/versioned"
	informers "k8s.io/delay-job-controller/pkg/generated/informers/externalversions"
	"k8s.io/delay-job-controller/pkg/protocol"
	"k8s.io/delay-job-controller/pkg/server"
	"k8s.io/delay-job-controller/pkg/utils"
	"k8s.io/delay-job-controller/version"
	"time"
)

func NewControllerServer() server.Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &ControllerServer{
		ctx:    ctx,
		cancel: cancel,
		stop:   make(chan struct{}, 2),
	}
}

type ControllerServer struct {
	ctx                context.Context
	cancel             context.CancelFunc
	delayJobController *delayjob.DelayJobController
	stop               chan struct{}
}

func (c *ControllerServer) Init() error {
	if err := config.Init(); err != nil {
		return err
	}
	if err := kubernetes.InitKubeClient(config.GetKubeCfg()); err != nil {
		return fmt.Errorf("ControllerServer.InitKubeClient failed cause: %s", err.Error())
	}
	return nil
}

func (c *ControllerServer) Start() error {
	log.Logger().Info("server starting")
	utils.Wrap(func() {
		log.Logger().Info("pprof/metrics service start addr: %s", config.GetCfg().PprofPort)
		if err := protocol.NewPprofMetricServer(config.GetCfg().PprofPort); err != nil {
			log.Logger().Error("ControllerServer.start NewPprofMetricServer error cause=%s", err.Error())
		}
	})
	delaySet, err := clientset.NewForConfig(config.GetKubeCfg())
	if err != nil {
		return fmt.Errorf("ControllerServer.Start NewDelayController failed cause=%s", err.Error())
	}
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubernetes.KubernetesClient(), time.Second*30)
	delayInformerFactory := informers.NewSharedInformerFactory(delaySet, time.Second*30)

	ctr, err := delayjob.NewController(
		kubernetes.KubernetesClient(),
		delaySet,
		kubeInformerFactory.Batch().V1().Jobs(),
		delayInformerFactory.Delayjob().V1().DelayJobs(),
	)
	if err != nil {
		return fmt.Errorf("ControllerServer.Start NewController failed cause=%s", err.Error())
	}
	kubeInformerFactory.Start(c.stop)
	delayInformerFactory.Start(c.stop)
	utils.Wrap(func() {
		if err = ctr.Run(c.ctx, 1); err != nil {
			log.Logger().Error("Error running controller: %s", err.Error())
		}
	})
	log.Logger().Info("version info: %+v", version.Version)
	log.Logger().Info("server started")
	return nil
}

func (c *ControllerServer) Stop() error {
	log.Logger().Info("server ready to close")
	c.cancel()
	c.stop <- struct{}{}
	c.stop <- struct{}{}
	var errs error
	log.Logger().Info("server is closed")
	log.Logger().Close()
	return errs
}
