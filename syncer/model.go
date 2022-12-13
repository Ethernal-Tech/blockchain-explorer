package syncer

import "sync"

var lock = &sync.Mutex{}

type signalSynch struct {
	Done chan struct{}
}

var synchInstance *signalSynch

func GetSignalSynchInstance() *signalSynch {
	if synchInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if synchInstance == nil {
			synchInstance = &signalSynch{
				Done: make(chan struct{}, 1),
			}
		}
	}
	return synchInstance
}
