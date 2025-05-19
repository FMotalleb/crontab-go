package event

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/core/utils"
)

func init() {
	eg.Register(newDockerGenerator)
}

func newDockerGenerator(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if cfg.Docker != nil {
		d := cfg.Docker
		con := utils.FirstNonZeroForced(d.Connection,
			"unix:///var/run/docker.sock",
		)
		return NewDockerEvent(
			con,
			d.Name,
			d.Image,
			d.Actions,
			d.Labels,
			utils.FirstNonZeroForced(d.ErrorLimit, 1),
			utils.FirstNonZeroForced(d.ErrorLimitPolicy, config.ErrorPolReconnect),
			utils.FirstNonZeroForced(d.ErrorThrottle, time.Second*5),
			log,
		)
	}
	return nil
}

type DockerEvent struct {
	connection       string
	containerMatcher regexp.Regexp
	imageMatcher     regexp.Regexp
	actions          *utils.List[events.Action]
	labels           map[string]regexp.Regexp
	errorThreshold   uint
	errorPolicy      config.ErrorLimitPolicy
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
	errorPolicy config.ErrorLimitPolicy,
	errorThrottle time.Duration,
	logger *logrus.Entry,
) abstraction.EventGenerator {
	return &DockerEvent{
		connection:       connection,
		containerMatcher: *regexp.MustCompile(containerMatcher),
		imageMatcher:     *regexp.MustCompile(imageMatcher),
		actions:          toAction(actions),
		labels:           reshapeLabelMatcher(labels),
		errorThreshold:   errorLimit,
		errorPolicy:      errorPolicy,
		errorThrottle:    errorThrottle,
		log:              logger,
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (de *DockerEvent) BuildTickChannel() <-chan []string {
	notifyChan := make(chan []string)

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
					case config.ErrorPolGiveUp:
						de.log.Warnf("Received more than %d consecutive errors from docker, marking instance as unstable and giving up this instance, no events will be received anymore", errs)
						close(notifyChan)
						return
					case config.ErrorPolKill:
						de.log.Fatalf("Received more than %d consecutive errors from docker, marking instance as unstable and killing in return, this may happen due to dockerd restarting", errs)
					case config.ErrorPolReconnect:
						de.log.Warnf("Received more than %d consecutive errors from docker, marking instance as unstable and retry connecting to docker", errs)
						for e := range de.BuildTickChannel() {
							notifyChan <- e
						}
					default:
						de.log.Fatalf("unexpected event.ErrorLimitPolicy: %#v, valid options are (kill,giv-up,reconnect)", de.errorPolicy)
					}
					errCount.Set(0)
				}
				if de.errorThrottle > 0 {
					time.Sleep(de.errorThrottle)
				}
			case event := <-msg:
				de.log.Trace("received an event from docker: ", event)
				if de.matches(&event) {
					notifyChan <- []string{"docker", event.Scope, string(event.Action), event.Actor.ID, strings.Join(reshapeAttrib(event.Actor.Attributes), ",")}
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

func reshapeAttrib(input map[string]string) []string {
	res := make([]string, 0, len(input))
	for k, v := range input {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

func toAction(acts []string) *utils.List[events.Action] {
	actions := utils.NewList[events.Action]()
	for _, act := range acts {
		actions.Add(events.Action(act))
	}
	return actions
}
