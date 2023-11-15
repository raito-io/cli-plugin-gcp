package syncer

import "github.com/google/uuid"

type IdGenerator struct {
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{}
}

func (i *IdGenerator) New() string {
	return uuid.NewString()
}
