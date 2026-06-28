package frontend

type Listable interface {
	GetTitle() string
	GetNumberingUrlString() string
	GetTimeRequiredString() string
}
