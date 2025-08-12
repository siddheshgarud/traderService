package consumer

import (
    "context"
    "encoding/json"
    "errors"
    "exchange-events-processor/pkg/config"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/segmentio/kafka-go"
)

/* =========================================================
   =============== External Kafka Consumer Lib =============
   ========================================================= */

type Health struct {
    ConnectionState bool     `json:"connection_state"`
    Brokers         []string `json:"brokers"`
    Topic           string   `json:"topic"`
    GroupID         string   `json:"group_id"`
    LastError       string   `json:"last_error,omitempty"`
}

type Message struct {
    Topic     string
    Partition int
    Offset    int64
    Key       []byte
    Value     []byte
    Timestamp time.Time
    Headers   []kafka.Header
}

type Config struct {
    Brokers        []string
    Address        []string
    Topics         []string
    GroupID        string
    Topic          string
    MinBytes       int
    MaxBytes       int
    CommitInterval time.Duration
    SASLUser       string
    SASLPassword   string
    Parallelism    int

    Timeout        time.Duration
    SessionTimeout time.Duration
    OffsetsInitial int
    CommitMode     int
    ConsumerMode   int
    RetryCount     int64

    NotificationHandler func(notification string)
    ErrorHandler        func(ctx context.Context, err error, message *Message)
    MessageHandler      func(ctx context.Context, message Message) error
}

// NewConfig returns a Kafka consumer config with default values
func NewConfig() *Config {
    return &Config{
        Timeout:        30 * time.Second,
        SessionTimeout: 10 * time.Second,
        OffsetsInitial: OffsetNewest,
        CommitMode:     OnMessageCompletion,
        ConsumerMode:   PullOrdered,
        RetryCount:     3,
    }
}

type KafkaConsumer struct {
    config   *Config
    health   Health
    mu       sync.RWMutex
    stopChan chan struct{}
    running  bool
    reader   *kafka.Reader
}

func New(config *Config) (*KafkaConsumer, error) {
    if config == nil {
        return nil, errors.New("invalid config")
    }
    if len(config.Brokers) == 0 || len(config.Topics) == 0 {
        return nil, errors.New("missing brokers or topics")
    }

    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:     config.Brokers,
        GroupID:     config.GroupID,
        Topic:       config.Topics[0], 
        MinBytes:    1,
        MaxBytes:    10e6,
        StartOffset: kafka.FirstOffset,
    })

    return &KafkaConsumer{
        config: config,
        health: Health{
            ConnectionState: true,
            Brokers:         config.Brokers,
            Topic:           config.Topics[0],
            GroupID:         config.GroupID,
        },
        stopChan: make(chan struct{}),
        reader:   r,
    }, nil
}

func (k *KafkaConsumer) Pull(ctx context.Context, handler func(Message) error) {
    k.mu.Lock()
    k.running = true
    k.mu.Unlock()
	log.Print("fetching message")
    go func() {
        defer k.reader.Close()
        for {
            select {
            case <-ctx.Done():
                return
            case <-k.stopChan:
                return
            default:
                m, err := k.reader.FetchMessage(ctx)
                if err != nil {
                    if k.config.ErrorHandler != nil {
                        k.config.ErrorHandler(ctx, err, nil)
                    }
                    continue
                }
                msg := Message{
                    Topic:     m.Topic,
                    Partition: m.Partition,
                    Offset:    m.Offset,
                    Key:       m.Key,
                    Value:     m.Value,
                    Timestamp: m.Time,
                    Headers:   m.Headers,
                }
                if handler != nil {
                    if err := handler(msg); err != nil && k.config.ErrorHandler != nil {
                        k.config.ErrorHandler(ctx, err, &msg)
                    }
                }
                _ = k.reader.CommitMessages(ctx, m)
            }
        }
    }()
}

func (k *KafkaConsumer) Close() error {
    k.mu.Lock()
    defer k.mu.Unlock()
    if k.running {
        close(k.stopChan)
        k.running = false
        k.health.ConnectionState = false
    }
    return nil
}

func (k *KafkaConsumer) Health() Health {
    k.mu.RLock()
    defer k.mu.RUnlock()
    return k.health
}

/* =========================================================
   ============== Application-Specific Logic ==============
   ========================================================= */

type Service interface {
    Pull(ctx context.Context) error
    Close() error
    Health() Health
    SetMessageHandler(topic string, handler func(Message) error)
}

type consumerService struct {
    consumer     *KafkaConsumer
    handlers     map[string]func(Message) error
    handlerMutex sync.RWMutex
}

func (s *consumerService) Pull(ctx context.Context) error {
    handler := func(msg Message) error {
        s.handlerMutex.RLock()
        h, ok := s.handlers[msg.Topic]
        s.handlerMutex.RUnlock()
        if ok && h != nil {
            return h(msg)
        }
        return nil
    }
    s.consumer.Pull(ctx, handler)
    return nil
}

func (s *consumerService) Close() error {
    return s.consumer.Close()
}

func (s *consumerService) Health() Health {
    return s.consumer.Health()
}

func (s *consumerService) SetMessageHandler(topic string, handler func(Message) error) {
    s.handlerMutex.Lock()
    s.handlers[topic] = handler
    s.handlerMutex.Unlock()
}

var GlobalConsumer Service

// OffsetNewest, OnMessageCompletion, PullOrdered are assumed constants
const (
    OffsetNewest        = 1
    OnMessageCompletion = 2
    PullOrdered         = 3
)

var Consumer *client

type client struct {
    log           log.Logger
    mu            sync.RWMutex
    handlers      map[string]func(ctx context.Context, message Message) error
    KafkaConsumer *KafkaConsumer
}

func LoadConsumer(ctx context.Context, log log.Logger) error {
    consumerConfig := NewConfig()

    consumerConfig.Brokers = config.KafkaConfig.Brokers
    consumerConfig.Topics = config.KafkaConfig.GetKafkaTopicNames()
    consumerConfig.GroupID = config.KafkaConfig.Consumer.Group
    consumerConfig.OffsetsInitial = OffsetNewest
    consumerConfig.CommitMode = OnMessageCompletion
    consumerConfig.ConsumerMode = PullOrdered
    consumerConfig.RetryCount = int64(config.KafkaConfig.Consumer.RetryCount)

    c := &client{
        log:      log,
        handlers: map[string]func(ctx context.Context, message Message) error{},
    }

    consumerConfig.NotificationHandler = c.notificationHandler
    consumerConfig.ErrorHandler = c.errorHandler
    consumerConfig.MessageHandler = c.messageHandler

    kafkaConsumer, err := New(consumerConfig)
    if err != nil {
        log.Fatal(ctx, "FailureCreatingKafkaConsumer", "Failed to create kafka consumer with error %v", err)
        return err
    }

    c.KafkaConsumer = kafkaConsumer
    Consumer = c

    // Wrap KafkaConsumer in consumerService for Service interface
    GlobalConsumer = &consumerService{
        consumer: kafkaConsumer,
        handlers: make(map[string]func(Message) error),
    }

    log.Printf("Constructed new kafka consumer successfully")
    return nil
}

func (c *client) notificationHandler(notification string) {
    c.log.Printf("Kafka consumer notification: %s", notification)
}

func (c *client) errorHandler(ctx context.Context, err error, message *Message) {
    if message == nil {
        c.log.Fatalf("Consumer-Error", "Error: %v", err)
    } else {
        c.log.Fatalf("Consumer-Error", "Error: %v, message at %s/%d --> %d", err, message.Topic, message.Partition, message.Offset)
    }
}

func (c *client) messageHandler(ctx context.Context, message Message) error {
    c.mu.RLock()
    defer c.mu.RUnlock()
    handler, ok := c.handlers[message.Topic]
    if !ok || handler == nil {
        return nil
    }
    return handler(ctx, message)
}

func (c *client) SetMessageHandler(topicName string, handler func(ctx context.Context, message Message) error) {
    c.mu.Lock()
    c.handlers[topicName] = handler
    c.mu.Unlock()
}

// Optional: HTTP health handler for debugging
func healthHandler(w http.ResponseWriter, r *http.Request) {
    if GlobalConsumer == nil {
        http.Error(w, "consumer not initialized", http.StatusServiceUnavailable)
        return
    }
    h := GlobalConsumer.Health()
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(h)
}