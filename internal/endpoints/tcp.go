package endpoints

type TCPEndpoint struct {
}

func (u *TCPEndpoint) Name() string {
	return "udp"
}

func (u *TCPEndpoint) Start(func()) {
}

func (u *TCPEndpoint) Stop() error {
	return nil
}
