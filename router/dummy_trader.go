package router

type DummyTrader struct {
	url     string
	profile *Profile
	err     error
}

func (t *DummyTrader) TradeSecret(secret string) (profile *Profile, err error) {
	profile = t.profile
	err = t.err
	return
}

func (t *DummyTrader) TradeName(name string) (profile *Profile, err error) {
	profile = t.profile
	err = t.err
	return
}

func (t *DummyTrader) Url() string {
	return t.url
}
