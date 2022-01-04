package stats

import (
	"lambdahttpgw/config"
)

type statsHolder struct {
	Hits       int64 `json:"hits"`
	LastReport int64 `json:"lastReport"`
}

var (
	hits  = map[string]*statsHolder{}
	hitCh chan string
)

// enableRecorder starts a goroutine that ensures single concurrency
// when writing to hits.
func enableRecorder() {
	hitCh = make(chan string)
	go func() {
		for true {
			functionName := <-hitCh
			record(functionName)
		}
	}()
}

func record(functionName string) {
	if !config.StatsReporterEnabled {
		return
	}
	holder, exist := hits[functionName]
	if !exist {
		holder = &statsHolder{
			Hits:       0,
			LastReport: 0,
		}
		hits[functionName] = holder
	}
	holder.Hits++
}

func RecordHit(functionName string) {
	hitCh <- functionName
}

func GetAllStats() map[string]*statsHolder {
	return hits
}
