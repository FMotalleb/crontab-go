package event

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/utils"
)

// TODO[epic=events] add watch method (probably after fs watcher is implemented)

func init() {
	registerGenerator(newLogListenerGenerator)
}

func newLogListenerGenerator(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if cfg.LogFile != "" {
		e, err := NewLogFile(
			cfg.LogFile,
			cfg.LogLineBreaker,
			cfg.LogMatcher,
			cfg.LogCheckCycle,
			log,
		)
		if err != nil {
			log.Error("Error creating LogFile: ", err)
			return nil
		}
		return e
	}
	return nil
}

// LogFile represents a log file that triggers an event when its content changes.
type LogFile struct {
	logger      *logrus.Entry
	filePath    string
	lineBreaker string
	matcher     regexp.Regexp
	// possibly will be deprecated or changed to an struct
	checkCycle time.Duration
}

// NewLogFile creates a new LogFile with the given parameters.
func NewLogFile(filePath string, lineBreaker string, matcherStr string, checkCycle time.Duration, logger *logrus.Entry) (*LogFile, error) {
	lineBreaker = utils.FirstNonZeroForced(lineBreaker, "\n")
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
func (lf *LogFile) BuildTickChannel() <-chan []string {
	notifyChan := make(chan []string)
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
		if err != nil && err != io.EOF {
			lf.logger.Warnln("error skipping initial data: ", err)
			return
		}
		for {
			data, err := reader.ReadString(byte(0))
			lf.logger.Tracef("read line: %s,Err: %s", data, err)
			if err != nil && err != io.EOF {
				lf.logger.Error("error reading log file: ", err)
				return
			}
			for _, line := range strings.Split(data, lf.lineBreaker) {
				eventData := []string{"log-file", lf.filePath, line}
				matches := lf.matcher.FindStringSubmatch(line)
				if matches != nil {
					names := lf.matcher.SubexpNames()

					eventData = append(eventData, reshapeRegxpMatch(names, matches)...)

					notifyChan <- eventData
				}
			}
			time.Sleep(lf.checkCycle)
		}
	}()
	return notifyChan
}

func reshapeRegxpMatch(keys []string, matches []string) []string {
	if len(keys) == 0 || len(matches) == 0 {
		return matches
	}
	result := make([]string, 0, len(keys))
	for i, key := range keys {
		if key != "" {
			result = append(result, fmt.Sprintf("%s=%s", key, matches[i]))
		} else if i == 0 {
			result = append(result, fmt.Sprintf("match=%s", matches[i]))
		}
	}
	return result
}
