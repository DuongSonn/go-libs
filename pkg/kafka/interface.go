package _kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

type IMessageProcessor interface {
	Process(ctx context.Context, msg *kgo.Record) error
}
