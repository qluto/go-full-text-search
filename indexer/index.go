package indexer

import ()

type Posting struct {
	DocId    string  `json:"docid"`
	Position int     `json:"position"`
	Tf       float64 `json:"tf"`
}

type PostingList []Posting

type InvRecord struct {
	Df          float64   `json:"df"`
	PostingList []Posting `json:"postinglist"`
}
