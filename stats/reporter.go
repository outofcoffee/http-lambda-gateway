package stats

import (
	"github.com/sirupsen/logrus"
	"lambdahttpgw/config"
	"time"
)

func Enable() chan bool {
	enableRecorder()

	statsUrl := config.GetStatsUrl()
	logrus.Debugf("enabling stats reporter to %s", statsUrl)

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
	logrus.Debugf("reporting stats")
}
