package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("URL not found")
	ErrURLExists     = errors.New("URL already exists")
	ErrAliasNoExists = errors.New("alias does not exist")
)
