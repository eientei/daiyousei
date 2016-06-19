package syncutil

import (
	"sync"
)

type SyncService interface {
	Running()    bool
	SubService() SyncService
	Stop()
	Wait()
	Done()
}

type syncServiceImpl struct {
	running  bool
	wgrp     *sync.WaitGroup
	parent   *syncServiceImpl
	children []*syncServiceImpl
}

func NewSyncService(parent SyncService) SyncService{
	var p *syncServiceImpl
	if parent != nil {
		p = parent.(*syncServiceImpl)
	}
	return &syncServiceImpl{
		running: true,
		wgrp: &sync.WaitGroup{},
		parent: p,
	}
}

func (svc *syncServiceImpl) Running() bool {
	return svc.running
}

func (svc *syncServiceImpl) SubService() SyncService {
	child := NewSyncService(svc)
	svc.children = append(svc.children, child.(*syncServiceImpl))
	svc.wgrp.Add(1)
	return child
}

func (svc *syncServiceImpl) Stop() {
	svc.running = false
	for _, c := range svc.children {
		c.Stop()
	}
}

func (svc *syncServiceImpl) Wait() {
	svc.wgrp.Wait()
}

func (svc *syncServiceImpl) Done() {
	svc.wgrp.Wait()
	if svc.parent != nil {
		svc.parent.wgrp.Done()
	}
}
/*
func NewSyncService(parent SyncService) SyncService {
	var p *syncServiceImpl
	if parent != nil {
		p = parent.(*syncServiceImpl)
	}
	svc := &syncServiceImpl{
		stopping: false,
		running: true,
		wgrp: &sync.WaitGroup{},
		parent: p,
	}
	//svc.wgrp.Add(1)
	return svc
}

func (svc *syncServiceImpl) Running() bool {
	return svc.running && !svc.stopping
}

func (svc *syncServiceImpl) SubService() SyncService {
	child := NewSyncService(svc)
	svc.wgrp.Add(1)
	logger.Debug(svc, "+1")
	svc.children = append(svc.children, child.(*syncServiceImpl))
	return child
}

func (svc *syncServiceImpl) Stop() {
	if svc.stopping {
		return
	}
	svc.stopping = true
	for _, c := range svc.children {
		c.Stop()
	}
	go func(){
		svc.wgrp.Wait()
		if svc.parent != nil {
			logger.Debug(svc.parent, "-1")
			svc.parent.wgrp.Done()
			for i, c := range svc.parent.children {
				if c == svc {
					l := len(svc.parent.children)-1
					svc.parent.children[i] = svc.parent.children[l]
					svc.parent.children = svc.parent.children[:l]
					break
				}
			}
		}

		svc.running = false
	}()
}

func (svc *syncServiceImpl) Wait() {
	if svc.running {
		svc.wgrp.Wait()
	}
}

func (svc *syncServiceImpl) Done() {
	if !svc.done {
		svc.done = true
		svc.Stop()
		logger.Debug(svc, "-1")
		svc.wgrp.Done()
	}
}*/