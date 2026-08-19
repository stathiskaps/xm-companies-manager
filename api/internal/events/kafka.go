package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

type CompanyEvent struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	CompanyID uuid.UUID `json:"company_id"`
	Timestamp time.Time `json:"timestamp"`
}

func NewProducer(brokers string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Producer{
		client: client,
		topic:  topic,
	}, nil
}

func (p *Producer) PublishCompanyEvent(
	ctx context.Context,
	eventType string,
	companyID uuid.UUID,
) error {
	event := CompanyEvent{
		ID:        uuid.New(),
		Type:      eventType,
		CompanyID: companyID,
		Timestamp: time.Now().UTC(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal company event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(companyID.String()),
		Value: value,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("publish company event: %w", err)
	}

	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
