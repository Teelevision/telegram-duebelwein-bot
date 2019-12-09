package medium

// Provider is the provider of a medium
type Provider interface {
	String() string
}

type simpleProvider string

func (p simpleProvider) String() string {
	return string(p)
}
