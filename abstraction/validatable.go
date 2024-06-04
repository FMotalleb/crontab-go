package abstraction

type (
	ValidatableStr string
	Validatable    interface {
		Validate() error
	}
)
