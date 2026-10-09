package services

const maxConcurrentCacheWrites = 16

var cacheWriteSlots = make(chan struct{}, maxConcurrentCacheWrites)

func runBoundedCacheWrite(task func()) bool {
	select {
	case cacheWriteSlots <- struct{}{}:
		go func() {
			defer func() { <-cacheWriteSlots }()
			task()
		}()
		return true
	default:
		return false
	}
}
