package abstraction

import "github.com/sirupsen/logrus"

type (
	Validatable interface {
		Validate(log *logrus.Entry) error
	}
)
