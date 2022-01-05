package stats

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"lambdahttpgw/config"
	"net/http"
	"strings"
	"time"
)

func Enable() chan bool {
	enableRecorder()

	logrus.Debugf("enabling stats reporter to %s", config.StatsUrl)

	done := make(chan bool)
	ticker := time.NewTicker(config.GetStatsInterval())

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				reportStats()
			}
		}
	}()

	return done
}

func reportStats() {
	logrus.Tracef("checking for pending stats")

	var pending = map[string]*statsHolder{}
	for funcName, holder := range GetAllStats() {
		if holder.Hits > holder.LastReport {
			pending[funcName] = holder
		}
	}
	if len(pending) == 0 {
		logrus.Tracef("no pending stats to report")
		return
	}

	logrus.Debugf("reporting %d pending stats", len(pending))
	for funcName, holder := range pending {
		// cache hits to enable background update
		hits := holder.Hits
		due := hits - holder.LastReport
		if due <= 0 {
			continue
		}
		if success := sendStat(funcName, due); success {
			holder.LastReport = hits
		}
	}
	logrus.Debugf("reported %d pending stats", len(pending))
}

func sendStat(funcName string, amount int64) bool {
	url := fmt.Sprintf("%s/hits/%s", config.StatsUrl, funcName)
	reqBody := strings.NewReader(fmt.Sprintf("%d", amount))
	request, err := http.NewRequest("PUT", url, reqBody)
	if err != nil {
		logrus.Warnf("failed to report stats for %s to %s: %s", funcName, url, err)
		return false
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logrus.Warnf("failed to report stats for %s to %s: %s", funcName, url, err)
		return false
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		logrus.Warnf("failed to report stats for %s to %s - received status code: %d", funcName, url, response.StatusCode)
		return false
	}
	logrus.Tracef("reported stats for %s to %s - received status code: %d", funcName, url, response.StatusCode)
	return true
}
