package extractor

import (
	"sync"
)

var extractors = make(map[string]extractor)
var mu sync.RWMutex

type extractor func(closer string) []Result
type Result struct {
	Name    string
	Class   string
	Time    string
	Message string
}

func Extract(key string, s string) []Result {
	mu.RLock()
	extractor := extractors[key]
	mu.RUnlock()
	return extractor(s)
}
func addExtractor(key string, extractor extractor) {
	mu.Lock()
	defer mu.Unlock()
	extractors[key] = extractor
}
