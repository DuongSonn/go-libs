package _kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

// tp represents a topic-partition pair for mapping consumers
type tp struct {
	t string
	p int32
}

// pconsumer handles consuming messages for a specific topic partition
type pconsumer struct {
	cl        *kgo.Client
	topic     string
	partition int32
	service   IMessageProcessor // Service that processes messages from this partition

	quit chan struct{}      // Channel to signal consumer to stop
	done chan struct{}      // Channel to signal consumer has stopped
	recs chan []*kgo.Record // Channel for passing records to be processed
}

// splitConsume manages multiple partition consumers
type splitConsume struct {
	// Using BlockRebalanceOnCommit means we do not need a mu to manage
	// consumers, unlike the autocommit normal example.
	consumers map[tp]*pconsumer
	service   IMessageProcessor
}

// consume processes messages from a specific partition
// This runs in its own goroutine for each partition
func (pc *pconsumer) consume() {
	defer close(pc.done)
	fmt.Printf("Starting consume for t %s p %d\n", pc.topic, pc.partition)
	defer fmt.Printf("Closing consume for t %s p %d\n", pc.topic, pc.partition)
	for {
		select {
		case <-pc.quit:
			return
		case recs := <-pc.recs:
			for _, rec := range recs {
				if err := pc.service.Process(context.Background(), rec); err != nil {
					fmt.Printf("Error when processing message err: %v t: %s p: %d offset %d\n", err, pc.topic, pc.partition, rec.Offset+1)
				}
			}
			err := pc.cl.CommitRecords(context.Background(), recs...)
			if err != nil {
				fmt.Printf("Error when committing offsets to kafka err: %v t: %s p: %d offset %d\n", err, pc.topic, pc.partition, recs[len(recs)-1].Offset+1)
			}
		}
	}
}

// assigned is called when partitions are assigned to this consumer
// It creates a new pconsumer for each assigned partition
func (s *splitConsume) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {
			pc := &pconsumer{
				cl:        cl,
				topic:     topic,
				partition: partition,
				service:   s.service,

				quit: make(chan struct{}),
				done: make(chan struct{}),
				recs: make(chan []*kgo.Record, 5),
			}
			s.consumers[tp{topic, partition}] = pc
			go pc.consume()
		}
	}
}

// lost is called when partitions are lost or revoked
// It stops the corresponding pconsumer instances
func (s *splitConsume) lost(_ context.Context, cl *kgo.Client, lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, partition := range partitions {
			tp := tp{topic, partition}
			pc := s.consumers[tp]
			delete(s.consumers, tp)
			close(pc.quit)
			fmt.Printf("waiting for work to finish t %s p %d\n", topic, partition)
			wg.Add(1)
			go func() { <-pc.done; wg.Done() }()
		}
	}
}

// poll continuously polls for records and distributes them to the appropriate partition consumers
func (s *splitConsume) poll(cl *kgo.Client) {
	for {
		// PollRecords is strongly recommended when using
		// BlockRebalanceOnPoll. You can tune how many records to
		// process at once (upper bound -- could all be on one
		// partition), ensuring that your processor loops complete fast
		// enough to not block a rebalance too long.
		fetches := cl.PollRecords(context.Background(), 10000)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(_ string, _ int32, err error) {
			// Note: you can delete this block, which will result
			// in these errors being sent to the partition
			// consumers, and then you can handle the errors there.
			panic(err)
		})
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			tp := tp{p.Topic, p.Partition}

			// Since we are using BlockRebalanceOnPoll, we can be
			// sure this partition consumer exists:
			//
			// * onAssigned is guaranteed to be called before we
			// fetch offsets for newly added partitions
			//
			// * onRevoked waits for partition consumers to quit
			// and be deleted before re-allowing polling.
			s.consumers[tp].recs <- p.Records
		})
		cl.AllowRebalance()
	}
}
