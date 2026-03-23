package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/example/messages-relay/internal/config"
	"github.com/example/messages-relay/internal/logging"
	"github.com/example/messages-relay/internal/mqtt"
	"github.com/example/messages-relay/internal/relay"
	"github.com/example/messages-relay/internal/validator"
)

func main() {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		panic(err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic("load config: " + err.Error())
	}

	// Logging
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, "Library", "Logs", "messages-relay")
	lgCfg := logging.Config{
		Level:      logging.DefaultLevel(),
		LogToFile:  true,
		LogDir:     logDir,
		LogFile:    "messages-relay.log",
		JSONOutput: true,
	}
	logger, closer, err := logging.NewLogger(lgCfg)
	if err != nil {
		panic("setup logging: " + err.Error())
	}
	if closer != nil {
		defer closer.Close()
	}

	v := validator.New(cfg)
	r := relay.New(cfg)

	handler := func(raw []byte) error {
		msg, err := v.Validate(raw)
		if err != nil {
			logger.Warn("validation failed", "error", err.Error())
			return err
		}
		if err := r.Send(msg.Destination, msg.Payload); err != nil {
			logger.Error("relay failed", "destination", msg.Destination, "error", err.Error())
			return err
		}
		logger.Info("message relayed", "destination", msg.Destination)
		return nil
	}

	mqttClient := mqtt.New(cfg, handler)
	if err := mqttClient.Connect(); err != nil {
		logger.Error("mqtt connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer mqttClient.Disconnect()

	logger.Info("messages-relay started", "topic", cfg.MQTT.Topic)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logger.Info("shutting down")
}
