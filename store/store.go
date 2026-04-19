package store

import "project/parser"

type EntryStore interface {
	Save(entry parser.Entry) error
}
