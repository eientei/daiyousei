package syncutil

import (
	"sync"
)

type SyncLatch struct {
	Running  bool
	Group    sync.WaitGroup
	Parent   *SyncLatch
	Children []*SyncLatch
	Handlers []func()
}

func NewSyncLatch() *SyncLatch {
	return &SyncLatch{
		Running: true,
		Group: sync.WaitGroup{},
	}
}

func (latch *SyncLatch) SubLatch() (child *SyncLatch) {
	latch.Group.Add(1)
	child = &SyncLatch{
		Running: true,
		Group: sync.WaitGroup{},
		Parent: latch,
	}
	latch.Children = append(latch.Children, child)
	return child
}

func (latch *SyncLatch) callHandlers() {
	if latch.Running {
		latch.Running = false
		length := len(latch.Handlers) - 1
		for i := range latch.Handlers {
			latch.Handlers[length - i]()
		}
	}
}

func (latch *SyncLatch) Terminate() {
	latch.callHandlers()
	for _, c := range latch.Children {
		c.Terminate()
	}
}

func (latch *SyncLatch) Await() {
	latch.Group.Wait()
}

func (latch *SyncLatch) Complete() {
	latch.callHandlers()
	if latch.Parent != nil {
		latch.Parent.Group.Done()
	}
}

func (latch *SyncLatch) Handle(f func()) {
	latch.Handlers = append(latch.Handlers, f)
}