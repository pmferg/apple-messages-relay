package mqtt

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/example/messages-relay/internal/config"
)

// Handler is called for each received message with the raw payload.
type Handler func(rawPayload []byte) error

// Client wraps the MQTT client and subscription.
type Client struct {
	cfg     *config.Config
	handler Handler
	client  mqtt.Client
	mu      sync.Mutex
}

// New creates a new MQTT client (does not connect yet).
func New(cfg *config.Config, handler Handler) *Client {
	return &Client{
		cfg:     cfg,
		handler: handler,
	}
}

// Connect establishes the MQTT connection and subscribes to the topic.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.cfg.MQTT.Broker)
	opts.SetUsername(c.cfg.MQTT.Username)
	opts.SetPassword(c.cfg.MQTT.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetOrderMatters(false)

	// TLS for ssl:// and mqtts://
	if isTLS(c.cfg.MQTT.Broker) {
		opts.SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})
	}

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		token := client.Subscribe(c.cfg.MQTT.Topic, 1, func(client mqtt.Client, msg mqtt.Message) {
			c.handler(msg.Payload())
		})
		token.Wait()
	})

	c.client = mqtt.NewClient(opts)
	token := c.client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt connect: %w", err)
	}
	return nil
}

// Disconnect closes the MQTT connection.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		c.client.Disconnect(250)
		c.client = nil
	}
}

func isTLS(broker string) bool {
	u, err := url.Parse(broker)
	if err != nil {
		return false
	}
	return u.Scheme == "ssl" || u.Scheme == "mqtts" || u.Scheme == "wss"
}
