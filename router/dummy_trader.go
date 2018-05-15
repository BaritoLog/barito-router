package router

type DummyTrader struct {
	url  string
	item *Profile
	err  error
}

func (t *DummyTrader) Trade(secret string) (item *Profile, err error) {
	item = t.item
	err = t.err
	return
}

func (t *DummyTrader) Url() string {
	return t.url
}
