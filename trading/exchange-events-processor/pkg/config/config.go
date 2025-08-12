package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"strings"
)

const (
	appServerPropertiesEnvKey      = "APP_SERVER_PROPERTIES"
	appKafkaConfigPropertiesEnvKey = "APP_KAFKA_PROPERTIES"
)

var (

	//KafkaConfig contains kafka configuration
	KafkaConfig kafkaConfig

	// AppConfig contains application configuration
	ServerConfig serverConfiguration
)

type (
	kafkaConfig struct {
		Brokers          []string              `json:"brokers"`
		OrderTopic       string                `json:"orderTopic"`
		TradesTopic      string                `json:"tradesTopic"`
		CashBalanceTopic string                `json:"cashBalanceTopic"`
		Consumer         kafkaConsumerSettings `json:"consumer"`
	}

	kafkaConsumerSettings struct {
		Group      string `json:"group"`
		RetryCount int    `json:"retryCount"`
	}

	serverConfiguration struct {
		Port 			int 	`json:"port"`
		UnmaskErrors	bool 	`json:"unmaskErrors"`
	}
)

func Load() error {

	//load app config
	if err := loadAppConfig(); err != nil {
		return err
	}

	//load kafka config
	if err := loadKafkaConfig(); err != nil {
		return err
	}
	return nil
}

func loadKafkaConfig() error {
	err := loadFromEnv(appKafkaConfigPropertiesEnvKey, &KafkaConfig)
	if err != nil {
		return err
	}

	if strings.TrimSpace(KafkaConfig.OrderTopic) == "" || strings.TrimSpace(KafkaConfig.TradesTopic) == "" || strings.TrimSpace(KafkaConfig.CashBalanceTopic) == "" || len(KafkaConfig.Brokers) == 0 || strings.TrimSpace(KafkaConfig.Consumer.Group) == "" {
		return errors.New("kafka config is missing")
	}

	return nil
}

func loadAppConfig() error {
	err := loadFromEnv(appServerPropertiesEnvKey, &ServerConfig)
	if err != nil {
		return err
	}

	return nil
}

func loadFromEnv(key string, cfg interface{}) error {
	v := os.Getenv(key)
	err := json.Unmarshal([]byte(v), cfg)
	if err != nil {
		return fmt.Errorf("getting env error from %v", key)
	}
	return nil
}


func (k kafkaConfig) GetKafkaTopicNames() []string{
	topics := []string{k.CashBalanceTopic , k.TradesTopic , k.OrderTopic}
	return topics
}