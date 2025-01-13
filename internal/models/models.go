package models

type Rates struct {
	Timestamp int64  `json:"timestamp"`
	AskPrice  string `json:"ask_price"`
	BidPrice  string `json:"bid_price"`
}
