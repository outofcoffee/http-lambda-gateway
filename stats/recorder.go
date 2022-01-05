package stats

import (
	"lambdahttpgw/config"
)

type statsHolder struct {
	Hits       int64 `json:"hits"`
	LastReport int64 `json:"lastReport"`
}

var (
	functionStats = map[string]*statsHolder{}
	hitCh         chan string
)

// enableRecorder starts a goroutine that ensures single concurrency
// when mutating functionStats.
func enableRecorder() {
	logrus.Debugf("enabling stats recorder")

	// buffer to reduce likelihood of blocking caller
	hitCh = make(chan string, 100)
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
	holder, exist := functionStats[functionName]
	if !exist {
		holder = &statsHolder{
			Hits:       0,
			LastReport: 0,
		}
		functionStats[functionName] = holder
	}
	holder.Hits++
}

func RecordHit(functionName string) {
	hitCh <- functionName
}

func GetAllStats() map[string]*statsHolder {
	return functionStats
}
