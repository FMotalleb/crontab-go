package event

import (
	"context"
	"regexp"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/core/utils"
)

type ErrorLimitPolicy string

const (
	Kill      ErrorLimitPolicy = "kill"
	GiveUp    ErrorLimitPolicy = "give-up"
	Reconnect ErrorLimitPolicy = "reconnect"
)

type DockerEvent struct {
	connection       string
	containerMatcher regexp.Regexp
	imageMatcher     regexp.Regexp
	actions          *utils.List[events.Action]
	labels           map[string]regexp.Regexp
	errorThreshold   uint
	errorPolicy      ErrorLimitPolicy
	errorThrottle    time.Duration
	log              *logrus.Entry
}

func NewDockerEvent(
	connection string,
	containerMatcher string,
	imageMatcher string,
	actions []string,
	labels map[string]string,
	errorLimit uint,
	errorPolicy ErrorLimitPolicy,
	errorThrottle time.Duration,
	logger *logrus.Entry,
) *DockerEvent {
	return &DockerEvent{
		connection:       connection,
		containerMatcher: *regexp.MustCompile(containerMatcher),
		imageMatcher:     *regexp.MustCompile(imageMatcher),
		actions:          toAction(logger, actions),
		labels:           reshapeLabelMatcher(labels),
		errorThreshold:   errorLimit,
		errorPolicy:      errorPolicy,
		errorThrottle:    errorThrottle,
		log:              logger,
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (de *DockerEvent) BuildTickChannel() <-chan any {
	notifyChan := make(chan any)

	cli, err := client.NewClientWithOpts(
		client.WithHost(de.connection),
	)
	if err != nil {
		de.log.Warn("failed to connect to docker: ", err)
		return de.BuildTickChannel()
	}
	go func() {
		ctx := context.Background()
		msg, err := cli.Events(ctx, events.ListOptions{})
		errCount := concurrency.NewLockedValue(uint(0))
		for {
			select {
			case err := <-err:
				de.log.WithError(err).Warn("received an error from docker: ", err)
				if de.errorThreshold == 0 {
					continue
				}
				errs := errCount.Get() + 1
				errCount.Set(errs)

				if errs >= de.errorThreshold {
					switch de.errorPolicy {
					case GiveUp:
						de.log.Warnf("Received more than %d consecutive errors from docker, marking instance as unstable and giving up this instance, no events will be received anymore", errs)
						close(notifyChan)
						return
					case Kill:
						de.log.Fatalf("Received more than %d consecutive errors from docker, marking instance as unstable and killing in return, this may happen due to dockerd restarting", errs)
					case Reconnect:
						de.log.Fatalf("Received more than %d consecutive errors from docker, marking instance as unstable and retry connecting to docker", errs)
					default:
						de.log.Fatalf("unexpected event.ErrorLimitPolicy: %#v, valid options are (kill,giv-up,reconnect)", de.errorPolicy)
					}
				}
				if de.errorThrottle > 0 {
					time.Sleep(de.errorThrottle)
				}
			case event := <-msg:
				de.log.Trace("received an event from docker: ", event)
				if de.matches(&event) {
					notifyChan <- false
				}
				errCount.Set(0)
			}
		}
	}()

	return notifyChan
}

func (de *DockerEvent) matches(msg *events.Message) bool {
	if de.actions.IsNotEmpty() && !de.actions.Contains(msg.Action) {
		return false
	}
	if !de.containerMatcher.MatchString(msg.Actor.Attributes["name"]) {
		return false
	}

	if !de.imageMatcher.MatchString(msg.Actor.Attributes["image"]) {
		return false
	}

	for k, matcher := range de.labels {
		if attrib, ok := msg.Actor.Attributes[k]; !ok && !matcher.MatchString(attrib) {
			return false
		}
	}
	return true
}

func reshapeLabelMatcher(labels map[string]string) map[string]regexp.Regexp {
	res := make(map[string]regexp.Regexp)
	for k, v := range labels {
		res[k] = *regexp.MustCompile(v)
	}
	return res
}

func toAction(log *logrus.Entry, acts []string) *utils.List[events.Action] {
	actions := utils.NewList[events.Action]()
	for _, act := range acts {
		actions.Add(events.Action(act))
	}
	return actions
}
