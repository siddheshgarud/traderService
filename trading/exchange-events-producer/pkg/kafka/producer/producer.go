package producer

import (
	"context"
	"encoding/json"
	"errors"
	"exchange-events-producer/pkg/config"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/segmentio/kafka-go"
)

// ======== Static/constant variables ========

const (
	DefaultBatchSize = 1
	DefaultTimeout   = 10 * time.Second
)

// ======== Global variables ========

var (
	GlobalProducer Service
	Producer       *client
	producers      map[string]*client
)

// ======== Types ========

type (
	Health struct {
		ConnectionState bool     `json:"connection_state"`
		Brokers         []string `json:"brokers"`
		Topic           string   `json:"topic"`
		LastError       string   `json:"last_error,omitempty"`
	}

	ProducerMessage struct {
		Topic     string
		Key       []byte
		Value     []byte
		Timestamp time.Time
		Headers   []kafka.Header
	}

	Config struct {
		Brokers        []string
		Topic          string
		BatchSize      int
		Timeout        time.Duration
		ErrorHandler   func(ctx context.Context, err error, message *ProducerMessage)
		SuccessHandler func(ctx context.Context, message ProducerMessage)
	}

	KafkaProducer struct {
		config   *Config
		health   Health
		mu       sync.RWMutex
		writer   *kafka.Writer
		running  bool
		stopChan chan struct{}
	}

	Service interface {
		SendMessage(ctx context.Context, msg ProducerMessage) error
		Close() error
		Health() Health
		SetSuccessHandler(handler func(ctx context.Context, message ProducerMessage))
		SetErrorHandler(handler func(ctx context.Context, err error, message *ProducerMessage))
	}

	producerService struct {
		producer     *KafkaProducer
		success      func(ctx context.Context, message ProducerMessage)
		errorHandler func(ctx context.Context, err error, message *ProducerMessage)
		cfg          Config
	}

	client struct {
		log           *log.Logger
		mu            sync.RWMutex
		KafkaProducer *KafkaProducer
	}
)

// ======== Functions ========

func NewConfig() *Config {
	return &Config{
		BatchSize: DefaultBatchSize,
		Timeout:   DefaultTimeout,
	}
}

func New(config *Config) (*KafkaProducer, error) {
	if config == nil {
		return nil, errors.New("invalid config")
	}
	if len(config.Brokers) == 0 || config.Topic == "" {
		return nil, errors.New("missing brokers or topic")
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    config.BatchSize,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &KafkaProducer{
		config:   config,
		health:   Health{ConnectionState: true, Brokers: config.Brokers, Topic: config.Topic},
		writer:   w,
		stopChan: make(chan struct{}),
	}, nil
}

func (k *KafkaProducer) SendMessage(ctx context.Context, msg ProducerMessage) error {
	k.mu.RLock()
	defer k.mu.RUnlock()
	k.running = true

	maxAttempts := uint(config.HTTPRequestConfig.MaxRetryAttempts)
	baseDelay := time.Duration(config.HTTPRequestConfig.BaseDelay) * time.Millisecond

	retryOpts := []retry.Option{
		retry.Attempts(maxAttempts),
		retry.Delay(baseDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			if k.config.ErrorHandler != nil {
				k.config.ErrorHandler(ctx, err, &msg)
			}
		}),
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
	}
	return retry.Do(
		func() error {
			kafkaMsg := kafka.Message{
				Key:     msg.Key,
				Value:   msg.Value,
				Time:    msg.Timestamp,
				Headers: msg.Headers,
			}
			err := k.writer.WriteMessages(ctx, kafkaMsg)
			log.Print(err)
			if err == nil && k.config.SuccessHandler != nil {
				k.config.SuccessHandler(ctx, msg)
			}
			if err != nil {
				log.Print(err)
				k.health.LastError = err.Error()
			}
			return err
		},
		retryOpts...,
	)
}

func (k *KafkaProducer) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.running {
		select {
		case <-k.stopChan:
			// already closed, do nothing
		default:
			close(k.stopChan)
		}
		k.running = false
		k.health.ConnectionState = false
	}

	return k.writer.Close()
}

func (k *KafkaProducer) Health() Health {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.health
}

func LoadProducers(ctx context.Context, log *log.Logger) error {
	producers = make(map[string]*client)

	for _, topic := range []string{
		config.KafkaConfig.OrderTopic,
		config.KafkaConfig.CashBalanceTopic,
		config.KafkaConfig.TradesTopic,
	} {
		producerConfig := NewConfig()
		producerConfig.Brokers = config.KafkaConfig.Brokers
		producerConfig.Topic = topic
		c := &client{
			log: log,
		}
		producerConfig.ErrorHandler = c.errorHandler
		producerConfig.SuccessHandler = c.successHandler

		service, err := New(producerConfig)
		if err != nil {
			return err
		}
		c.KafkaProducer = service
		producers[topic] = c
	}
	return nil
}

func (c *client) errorHandler(ctx context.Context, err error, message *ProducerMessage) {
	if message == nil {
		c.log.Printf("Producer-Error: %v", err)
	} else {
		c.log.Printf("Producer-Error: %v, message to %s", err, message.Topic)
	}
}

func (c *client) successHandler(ctx context.Context, message ProducerMessage) {
	c.log.Printf("Message produced successfully to %s", message.Topic)
}

func ProduceWithRetry(ctx context.Context, msg ProducerMessage) error {
	p, ok := producers[msg.Topic]
	if !ok {
		return fmt.Errorf("no producer found for topic %s", msg.Topic)
	}
	return p.KafkaProducer.SendMessage(ctx, msg)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if GlobalProducer == nil {
		http.Error(w, "producer not initialized", http.StatusServiceUnavailable)
		return
	}
	h := GlobalProducer.Health()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h)
}
