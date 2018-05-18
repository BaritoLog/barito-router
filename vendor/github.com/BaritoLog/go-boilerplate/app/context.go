package app

type Context interface {
	Init(config Configuration) (err error)
	Run() (err error)
}

type dummyContext struct {
	initErr error
	runErr  error
}

func (c dummyContext) Init(config Configuration) error {
	return c.initErr
}
func (c dummyContext) Run() error {
	return c.runErr
}
