package task

import (
	"fmt"
	ext "github.com/hstreamdb/dev-deploy/pkg/executor"
	"github.com/hstreamdb/dev-deploy/pkg/service"
	"sync"
)

type HStoreClusterCtx struct {
	ctx     *service.GlobalCtx
	service []*service.HStore
}

type InitStoreEnv struct {
	HStoreClusterCtx
}

func (s *InitStoreEnv) String() string {
	return "Task: init store environment"
}

func (s *InitStoreEnv) Run(executor ext.Executor) error {
	return serviceInitEnv(executor, s.ctx, s.service)
}

type SyncStoreConfig struct {
	HStoreClusterCtx
}

func (s *SyncStoreConfig) String() string {
	return "Task: sync store config"
}

func (s *SyncStoreConfig) Run(executor ext.Executor) error {
	return configSync(executor, s.ctx, s.service)
}

type StartStoreCluster struct {
	HStoreClusterCtx
}

func (s *StartStoreCluster) String() string {
	return "Task: start store cluster"
}

func (s *StartStoreCluster) Run(executor ext.Executor) error {
	return serviceDeploy(executor, s.ctx, s.service)
}

type WaitStoreReady struct {
	HStoreClusterCtx
}

func (w *WaitStoreReady) String() string {
	return "Task: wait store ready"
}

func (w *WaitStoreReady) Run(executor ext.Executor) error {
	for _, store := range w.service {
		executorCtx := store.CheckReady(w.ctx)
		target := fmt.Sprintf("%s:%d", executorCtx.Target, w.ctx.SSHPort)
		res, err := executor.Execute(target, executorCtx.Cmd)
		if err != nil {
			return fmt.Errorf("%s-%s", err.Error(), res)
		}
	}
	return nil
}

type BootStrap struct {
	HStoreClusterCtx
}

func (b *BootStrap) String() string {
	return "Task: bootstrap"
}

func (b *BootStrap) Run(executor ext.Executor) error {
	var adminStore *service.HStore
	for _, store := range b.service {
		if store.IsAdmin() {
			adminStore = store
			break
		}
	}

	executorCtx := adminStore.Bootstrap(b.ctx)
	target := fmt.Sprintf("%s:%d", executorCtx.Target, b.ctx.SSHPort)
	res, err := executor.Execute(target, executorCtx.Cmd)
	if err != nil {
		return fmt.Errorf("%s-%s", err.Error(), res)
	}
	return nil
}

type RemoveStore struct {
	HStoreClusterCtx
}

func (r *RemoveStore) String() string {
	return "Task: remove store"
}

func (r *RemoveStore) Run(executor ext.Executor) error {
	return serviceRemove(executor, r.ctx, r.service)
}

type MountDisk struct {
	HStoreClusterCtx
}

func (m *MountDisk) String() string {
	return "Task: mount disk"
}

func (m *MountDisk) Run(executor ext.Executor) error {
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var firstErr error
	wg.Add(len(m.service))
	for _, svc := range m.service {
		go func(svc *service.HStore) {
			defer wg.Done()
			executorCtx := svc.MountDisk()
			target := fmt.Sprintf("%s:%d", executorCtx.Target, m.ctx.SSHPort)
			res, err := executor.Execute(target, executorCtx.Cmd)
			if err != nil {
				mutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s-%s", err.Error(), res)
				}
				mutex.Unlock()
			}
		}(svc)
	}
	wg.Wait()
	return firstErr
}