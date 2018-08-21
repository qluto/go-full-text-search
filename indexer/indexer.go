package indexer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	_ "regexp"
	_ "strconv"
	"strings"
)

func tokenize(s string) []string {
	// TODO: implement filtering out stop words
	// TODO: implement stemming for conjugation
	s = strings.Replace(s, ",", "", -1)
	s = strings.Replace(s, ".", "", -1)
	return strings.Split(s, " ") // tokenize english sentences
}

func uniqCount(tokens []string) map[string]int {
	uniq := make(map[string]int)
	for _, v := range tokens {
		num, ok := uniq[v]
		if ok {
			uniq[v] = num + 1
		} else {
			uniq[v] = 1
		}
	}
	return uniq
}

func filter(pl []Posting, target string) ([]Posting, bool) {
	var ret []Posting
	isFiltered := false
	for _, p := range pl {
		if p.DocId != target {
			ret = append(ret, p)
		} else {
			isFiltered = true
		}
	}

	return ret, isFiltered
}

func DeleteDoc(docid string, db *bolt.DB) {
	db.Update(func(tx *bolt.Tx) error {
		invBucket, err := tx.CreateBucketIfNotExists([]byte("invindex"))
		if err != nil {
			log.Fatal(err)
		}
		c := invBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			invRecord := new(InvRecord)
			if err := json.Unmarshal(v, invRecord); err != nil {
				log.Fatal(err)
			}
			pl, isFiltered := filter(invRecord.PostingList, docid)
			if isFiltered {
				invRecord.Df -= 1
			}
			invRecord.PostingList = pl

			jsonBytes, err := json.Marshal(invRecord)
			if err != nil {
				log.Fatal(err)
			}
			err = invBucket.Put([]byte(k), jsonBytes)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
}

func AddDoc(docid string, body string, db *bolt.DB) {
	tokens := tokenize(body)
	uniq := uniqCount(tokens)

	db.Update(func(tx *bolt.Tx) error {

		invBucket, err := tx.CreateBucketIfNotExists([]byte("invindex"))
		if err != nil {
			log.Fatal(err)
		}
		// すでに存在するドキュメントの転置インデックス更新の場合は既存のを削除してから処理する
		c := invBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			invRecord := new(InvRecord)
			if err := json.Unmarshal(v, invRecord); err != nil {
				log.Fatal(err)
			}
			pl, isFiltered := filter(invRecord.PostingList, docid)
			if isFiltered {
				invRecord.Df -= 1
			}
			invRecord.PostingList = pl

			jsonBytes, err := json.Marshal(invRecord)
			if err != nil {
				log.Fatal(err)
			}
			err = invBucket.Put([]byte(k), jsonBytes)
			if err != nil {
				log.Fatal(err)
			}
		}

		for i, v := range tokens {
			record := invBucket.Get([]byte(v))
			if record == nil {
				tf := float64(uniq[v]) / float64(len(tokens))
				var invRecord = &InvRecord{
					Df:          1,
					PostingList: []Posting{{DocId: docid, Position: i, Tf: tf}},
				}
				jsonBytes, err := json.Marshal(invRecord)
				if err != nil {
					log.Fatal(err)
				}
				err = invBucket.Put([]byte(v), jsonBytes)
			} else {
				invRecord := new(InvRecord)
				if err := json.Unmarshal(record, invRecord); err != nil {
					log.Fatal(err)
				}
				tf := float64(uniq[v]) / float64(len(tokens))
				isAlreadyExists := false
				for _, v := range invRecord.PostingList {
					if v.DocId == docid {
						isAlreadyExists = true
					}
				}
				invRecord.PostingList = append(invRecord.PostingList, Posting{DocId: docid, Position: i, Tf: tf})
				if !isAlreadyExists {
					invRecord.Df += 1
				}
				jsonBytes, err := json.Marshal(invRecord)
				if err != nil {
					log.Fatal(err)
				}
				err = invBucket.Put([]byte(v), jsonBytes)
			}
		}
		if err != nil {
			log.Fatal(err)
		}

		docsBucket, err := tx.CreateBucketIfNotExists([]byte("docs"))
		if err != nil {
			log.Fatal(err)
		}
		target := docsBucket.Get([]byte(docid))
		err = docsBucket.Put([]byte(docid), []byte(body))
		if err != nil {
			log.Fatal(err)
		}

		statBucket, err := tx.CreateBucketIfNotExists([]byte("stat"))
		docNumBytes := statBucket.Get([]byte("docNum"))
		var num uint64 = 0
		if docNumBytes != nil {
			if target != nil {
				num = Btoi(docNumBytes)
			} else {
				num = Btoi(docNumBytes) + 1
			}
		} else {
			num += 1
		}
		err = statBucket.Put([]byte("docNum"), Itob(num))
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})
}

func Itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func Btoi(b []byte) uint64 {
	padding := make([]byte, 8-len(b))
	i := binary.BigEndian.Uint64(append(padding, b...))
	return i
}

func DumpCursor(tx *bolt.Tx, c *bolt.Cursor, indent int) string {
	var dumpString []string
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			//fmt.Printf(strings.Repeat("\t", indent)+"[%s]\n", k)
			dumpString = append(dumpString, fmt.Sprintf(strings.Repeat("\t", indent)+"[%s]\n", k))
			newBucket := c.Bucket().Bucket(k)
			if newBucket == nil {
				newBucket = tx.Bucket(k)
			}
			newCursor := newBucket.Cursor()
			dumpString = append(dumpString, DumpCursor(tx, newCursor, indent+1))
		} else {
			dumpString = append(dumpString, fmt.Sprintf(strings.Repeat("\t", indent)+"%s\n", k))
			dumpString = append(dumpString, fmt.Sprintf(strings.Repeat("\t", indent+1)+"%s\n", v))
			//fmt.Printf(strings.Repeat("\t", indent)+"%s\n", k)
			//fmt.Printf(strings.Repeat("\t", indent+1)+"%s\n", v)
		}

	}
	return strings.Join(dumpString[:], "")
}
