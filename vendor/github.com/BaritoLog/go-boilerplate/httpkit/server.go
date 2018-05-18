package httpkit

type Server interface {
	ListenAndServe() error
}

type DummyServer struct {
	ErrListAndServer error
}

func (s DummyServer) ListenAndServe() error {
	return s.ErrListAndServer
}
