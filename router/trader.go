package router

type Trader interface {
	Trade(secret string) *Item
	Url() string
}

// NewHttpTrader
func NewTrader(url string) Trader {
	return &trader{url: url}
}

type trader struct {
	url string
}

// Address
func (t *trader) Url() string {
	return t.url
}

// Trade
func (t *trader) Trade(secret string) *Item {
	return &Item{}
}
