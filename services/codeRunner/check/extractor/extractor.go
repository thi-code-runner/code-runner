package extractor

import (
	"code-runner/model"
	"sync"
)

var extractors = make(map[string]extractor)
var mu sync.RWMutex

type extractor func(closer string) []*model.Detail

func Extract(key string, s string) []*model.Detail {
	mu.RLock()
	extractor, ok := extractors[key]
	mu.RUnlock()
	if !ok {
		return make([]*model.Detail, 0)
	}
	return extractor(s)
}
func addExtractor(key string, extractor extractor) {
	mu.Lock()
	defer mu.Unlock()
	extractors[key] = extractor
}
