package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	credential "github.com/FMotalleb/crontab-go/core/os_credential"
)

// Validate checks the validity of a Task.
// It ensures that the task has exactly one of the Get, Post, or Command fields, and validates other fields based on the specified action.
// If any validation fails, it returns an error with the specific validation error.
// Otherwise, it returns nil.
func (t *Task) Validate(log *logrus.Entry) error {
	// Log the start of validation
	log.Tracef("Validating Task: %+v", t)
	checkList := []func(*Task, *logrus.Entry) error{
		validateActionsList,
		validateCredential,
		validateFields,
		validateGetRequest,
		validateTimeout,
		validatePostData,
		validateRetry,
	}
	for _, check := range checkList {
		if err := check(t, log); err != nil {
			return err
		}
	}

	// Validate hooks
	for _, task := range append(t.OnDone, t.OnFail...) {
		if err := task.Validate(log); err != nil {
			joinedErr := errors.Join(errors.New("hook: failed to validate"), err)
			log.WithError(joinedErr).Warn("Validation failed for Task")
			return joinedErr
		}
	}

	// Log the successful validation
	log.Tracef("Validation successful for Task: %+v", t)
	return nil
}

func validateRetry(t *Task, log *logrus.Entry) error {
	if t.RetryDelay < 0 {
		err := fmt.Errorf(
			"retry delay for tasks cannot be negative received `%d` for %+v",
			t.RetryDelay,
			t,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}
	return nil
}

func validatePostData(t *Task, log *logrus.Entry) error {
	if t.Data != nil {
		_, err := json.Marshal(t.Data)
		if err != nil {
			log.WithError(err).Warn("Validation failed for Task")
			return err
		}
	}
	return nil
}

func validateTimeout(t *Task, log *logrus.Entry) error {
	if t.Timeout < 0 {
		err := fmt.Errorf(
			"timeout for tasks cannot be negative received `%d` for %+v",
			t.Timeout,
			t,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}
	return nil
}

func validateGetRequest(t *Task, log *logrus.Entry) error {
	if t.Get != "" && t.Data != nil {
		err := fmt.Errorf("GET request cannot have data field, violating GET URI: `%s`", t.Get)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}
	return nil
}

func validateFields(t *Task, log *logrus.Entry) error {
	if t.Command != "" && (t.Data != nil || t.Headers != nil) {
		err := fmt.Errorf("command cannot have data or headers field, violating command: `%s`", t.Command)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}
	return nil
}

func validateCredential(t *Task, log *logrus.Entry) error {
	if err := credential.Validate(log, t.UserName, t.GroupName); err != nil {
		log.WithError(err).Warn("Be careful when using credentials, in local mode you can't use credentials unless running as root")
	}
	return nil
}

func validateActionsList(t *Task, log *logrus.Entry) error {
	actions := []bool{
		t.Get != "",
		t.Command != "",
		t.Post != "",
	}
	activeActions := 0
	for _, t := range actions {
		if t {
			activeActions++
		}
	}
	if activeActions != 1 {
		err := fmt.Errorf(
			"a single task should have one of (Get, Post, Command) fields, received:(Command: `%s`, Get: `%s`, Post: `%s`)",
			t.Command,
			t.Get,
			t.Post,
		)
		log.WithError(err).Warn("Validation failed for Task")
		return err
	}
	return nil
}
