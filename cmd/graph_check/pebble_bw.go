package graph_check

import (
	"bytes"
	"log"

	"github.com/cockroachdb/pebble"
)

type pebbleBulkWrite struct {
	db              *pebble.DB
	batch           *pebble.Batch
	highest, lowest []byte
	curSize         int
}

const (
	maxWriterBuffer = 3 << 30
)

func copyBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func (pbw *pebbleBulkWrite) Set(id []byte, val []byte) error {
	//log.Printf("Setting %x %s", id, val)
	pbw.curSize += len(id) + len(val)
	if pbw.highest == nil || bytes.Compare(id, pbw.highest) > 0 {
		pbw.highest = copyBytes(id)
	}
	if pbw.lowest == nil || bytes.Compare(id, pbw.lowest) < 0 {
		pbw.lowest = copyBytes(id)
	}
	err := pbw.batch.Set(id, val, nil)
	if pbw.curSize > maxWriterBuffer {
		log.Printf("Running batch Commit")
		pbw.batch.Commit(nil)
		pbw.batch.Reset()
		pbw.curSize = 0
	}
	return err
}

func (pbw *pebbleBulkWrite) Close() error {
	log.Printf("Running batch close Commit")
	err := pbw.batch.Commit(nil)
	if err != nil {
		return err
	}
	pbw.batch.Close()
	if pbw.lowest != nil && pbw.highest != nil {
		pbw.db.Compact(pbw.lowest, pbw.highest, true)
	}
	return nil
}
