package options

type ServerRunOptions struct {
	InsecurePort int

	BindAddress string

	SecurePort int

	TlsCertFile string

	TlsPrivateKey string
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  9090,
		SecurePort:    0,
		TlsCertFile:   "",
		TlsPrivateKey: "",
	}

	return &s
}
