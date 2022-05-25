package utils

import "k8s.io/client-go/tools/cache"

func NewCatchControllerWrapper(c cache.Controller) *controllerWrapper {
	return &controllerWrapper{
		stop: make(chan struct{}, 1),
		c:    c,
	}
}

type controllerWrapper struct {
	stop chan struct{}
	c    cache.Controller
}

func (w *controllerWrapper) Run() {
	w.c.Run(w.stop)
}

func (w *controllerWrapper) Close() {
	w.stop <- struct{}{}
}
