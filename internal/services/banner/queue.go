package banner

import (
	"bytes"
	"encoding/gob"
)

type DeleteBannerTask struct {
	ID int64
}

type BannerQueue struct {
	queue     Queue
	queueName string
}

func NewBannerQueue(queue Queue, queueName string) *BannerQueue {
	return &BannerQueue{
		queue:     queue,
		queueName: queueName,
	}
}

func (q *BannerQueue) Publish(task *DeleteBannerTask) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(task); err != nil {
		return err
	}

	return q.queue.Publish(buf.Bytes(), q.queueName)
}

func (q *BannerQueue) Consume() (*DeleteBannerTask, error) {
	messages, err := q.queue.Consume(q.queueName)
	if err != nil {
		return nil, err
	}

	msg := <-messages
	var task DeleteBannerTask
	dec := gob.NewDecoder(bytes.NewBuffer(msg))
	if err := dec.Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}
