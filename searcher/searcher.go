package searcher

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	indexer "github.com/qluto/go-full-text-search/indexer"
	"log"
	"math"
	"sort"
	"strings"
)

type QueryTree interface {
	Accumulate(db *bolt.DB) Results
}

type IndexReader struct {
	q string
}

type AndOperator struct {
	Left  QueryTree
	Right QueryTree
}

type OrOperator struct {
	Left  QueryTree
	Right QueryTree
}

type PhraseOperator struct {
	Left  QueryTree
	Right QueryTree
}

type NotOperator struct {
	Left  QueryTree // NotOperator以外のOperators
	Right QueryTree // 除外条件
}

func (x IndexReader) Accumulate(db *bolt.DB) Results {
	rs := Results{}
	db.View(func(tx *bolt.Tx) error {
		statBucket := tx.Bucket([]byte("stat"))
		if statBucket == nil {
			log.Fatal("stat bucket not found")
		}
		v := statBucket.Get([]byte("docNum"))
		numOfDocs := indexer.Btoi(v)

		bucket := tx.Bucket([]byte("invindex"))
		if bucket == nil {
			log.Fatal("bucket not found")
		}
		v = bucket.Get([]byte(x.q))
		if v == nil {
			return nil
		}

		invRecord := new(indexer.InvRecord)
		if err := json.Unmarshal(v, invRecord); err != nil {
			log.Fatal(err)
		}
		m := make(map[string]bool)
		for _, p := range invRecord.PostingList {
			if !m[p.DocId] {
				m[p.DocId] = true
				score := p.Tf * (math.Log(float64(numOfDocs)/invRecord.Df) + 1)
				rs = append(rs, Result{p.DocId, score})
			}
		}

		return nil
	})
	return rs
}

func (x AndOperator) Accumulate(db *bolt.DB) Results {
	l := x.Left.Accumulate(db)
	r := x.Right.Accumulate(db)
	m := make(map[string]Result)
	rs := Results{}
	for _, v := range l {
		m[v.DocId] = v
	}
	for _, v := range r {
		if _, ok := m[v.DocId]; ok {
			m[v.DocId] = Result{v.DocId, v.Score + m[v.DocId].Score}
		} else {
			delete(m, v.DocId)
		}
	}
	for _, v := range m {
		rs = append(rs, v)
	}

	return rs
}

func (x OrOperator) Accumulate(db *bolt.DB) Results {
	l := x.Left.Accumulate(db)
	r := x.Right.Accumulate(db)
	m := make(map[string]Result)
	rs := Results{}
	for _, v := range l {
		m[v.DocId] = v
	}
	for _, v := range r {
		if _, ok := m[v.DocId]; ok {
			m[v.DocId] = Result{v.DocId, v.Score + m[v.DocId].Score}
		} else {
			m[v.DocId] = v
		}
	}
	for _, v := range m {
		rs = append(rs, v)
	}
	return rs
}

func (x PhraseOperator) Accumulate(db *bolt.DB) Results {
	// TODO AndOperatorの派生形で、token前後の距離が1の結果だけを返すようにする
	return Results{}
}

func (x NotOperator) Accumulate(db *bolt.DB) Results {
	// TODO
	return Results{}
}

type Result struct {
	DocId string
	Score float64
}

type Results []Result

func (rs Results) Len() int {
	return len(rs)
}

func (rs Results) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs Results) Less(i, j int) bool {
	return rs[i].Score > rs[j].Score
}

func parseQuery(q string) QueryTree {
	// TODO
	//return AndOperator{IndexReader{"Lorem"}, IndexReader{"amet"}}
	return IndexReader{}
}

func parseResults(rs Results) string {
	var resString []string
	for _, v := range rs {
		resString = append(resString, fmt.Sprintf("{\"docid\":\"%s\",\"score\":%f}", v.DocId, v.Score))
	}
	return "{[" + strings.Join(resString[:], ",") + "]}"
}

func Search(q string, db *bolt.DB) string {
	rs := parseQuery(q).Accumulate(db)
	sort.Sort(rs)
	return parseResults(rs)
}
