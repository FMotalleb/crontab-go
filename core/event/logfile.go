package event

import (
	"bufio"
	"io"
	"math"
	"os"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/core/utils"
)

// TODO[epic=events] add watch method (probably after fs watcher is implemented)

// LogFile represents a log file that triggers an event when its content changes.
type LogFile struct {
	logger      *logrus.Entry
	filePath    string
	lineBreaker rune
	matcher     regexp.Regexp
	// possibly will be deprecated or changed to an struct
	checkCycle time.Duration
}

// NewLogFile creates a new LogFile with the given parameters.
func NewLogFile(filePath string, lineBreaker rune, matcherStr string, checkCycle time.Duration, logger *logrus.Entry) (*LogFile, error) {
	lineBreaker = utils.FirstNonZeroForced(lineBreaker, '\n')
	matcherStr = utils.FirstNonZeroForced(matcherStr, ".")
	checkCycle = utils.FirstNonZeroForced(checkCycle, time.Second)

	matcher, err := regexp.Compile(
		matcherStr,
	)
	if err != nil {
		return nil, err
	}
	return &LogFile{
		logger: logger.WithFields(
			logrus.Fields{
				"scheduler":    "log_file",
				"file":         filePath,
				"line_breaker": lineBreaker,
				"matcher":      matcherStr,
				"check_cycle":  checkCycle,
			},
		),
		filePath:    filePath,
		lineBreaker: lineBreaker,
		matcher:     *matcher,
		checkCycle:  checkCycle,
	}, nil
}

// BuildTickChannel implements abstraction.Scheduler.
func (lf *LogFile) BuildTickChannel() <-chan any {
	notifyChan := make(chan any)
	go func() {
		// Use bufio to read file line by line
		file, err := os.Open(lf.filePath)
		if err != nil {
			lf.logger.Error("failed to open log file: ", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				lf.logger.Warnf("failed to close log file: %s", err)
			}
		}()
		reader := bufio.NewReader(file)
		_, err = reader.Discard(math.MaxInt64)
		if err != nil {
			lf.logger.Warnln("error skipping initial data: ", err)
			return
		}
		for {
			line, err := reader.ReadString(byte(lf.lineBreaker))
			lf.logger.Tracef("read line: %s,Err: %s", line, err)
			if err != nil && err != io.EOF {
				lf.logger.Error("error reading log file: ", err)
				return
			}

			if lf.matcher.MatchString(line) {
				notifyChan <- false
			}
			time.Sleep(lf.checkCycle)
		}
	}()
	return notifyChan
}
