package endpoints

// Endpoint represents implementation of entrypoint into a DNS resolver.
type Endpoint interface {
	Name() string
	Start(func())
	Stop() error
}
