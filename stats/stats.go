package stats

import (
	"github.com/sirupsen/logrus"
	"lambdahttpgw/config"
)

func Init() {
	if config.StatsRecorderEnabled {
		enableRecorder()
	} else {
		logrus.Debugf("stats recording is disabled")
	}
	if config.StatsReporterEnabled {
		enableReporter()
	} else {
		logrus.Debugf("stats reporting is disabled")
	}
}
