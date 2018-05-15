package router

type DummyTrader struct {
	url  string
	item *Item
	err  error
}

func (t *DummyTrader) Trade(secret string) (item *Item, err error) {
	item = t.item
	err = t.err
	return
}

func (t *DummyTrader) Url() string {
	return t.url
}
