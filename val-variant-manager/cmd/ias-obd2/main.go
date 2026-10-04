// © 2026 IAS Runner.
// Authored by Jordan Himawan (web.jojohimawan.cloud).
// Cyber Security Research Group, PENS.
// ----------------------------------------------------
// The illusion is half the disease,
// the reassurance is half the medicine,
// and patience is the first step of healing.
// —Ibn Sina.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Import the Core
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/config"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/decoding"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/kafka"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/mapper"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/pipeline"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/source"

	// Import the Protocols you want to enable

	obd2 "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// cfg gets the configuration values.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[ERR] Failed to log config: %v", err)
	}

	// Initialize infrastructure (Kafka, Mongo, Serial, CAN)
	mongoStore, err := mapper.NewMongoStore(ctx, mapper.MongoConfig{
		URI:        cfg.MongoDBURI,
		Database:   cfg.MongoDBDatabase,
		Collection: cfg.MongoDBCollection,
		Username:   cfg.MongoDBAuthUser,
		Password:   cfg.MongoDBAuthPassword,
		Timeout:    2 * time.Second,
	})
	if err != nil {
		log.Printf("[WARN] Mongo not available: %v", err)
	}

	mapperService := mapper.NewService(mongoStore, "./data/mappings.json")
	if err := mapperService.Init(context.Background()); err != nil {
		log.Printf("[WARN] Mapper init failed: %v", err)
	}

	producer, err := kafka.NewProducer(
		cfg.KafkaBrokerURL,
		cfg.SchemaRegistryURL,
	)
	if err != nil {
		log.Printf("[WARN] Kafka not available: %v", err)
	}
	defer producer.Close()

	serialReader, err := source.Open(cfg.SerialPort, 9600)
	if err != nil {
		log.Printf("[WARN] Serial GNSS not available: %v", err)
	}

	vcan, err := source.Connect(ctx, cfg.CanNetwork, cfg.CanNetworkAddress)
	if err != nil {
		log.Fatalf("Failed to connect to CAN interface %s: %v", cfg.CanNetworkAddress, err)
	}
	defer vcan.Close()

	// Initialize decoder manager.
	decoderManager := decoding.NewManager()

	// Register decoder plugins (modules).
	decoderManager.RegisterModule(obd2.New())
	log.Println("[INFO] Intelligent Agent System Runner Started...")
	log.Printf("[INFO] Loaded Protocols: %s", obd2.New().Name())

	// Start pipeline.
	runner := pipeline.NewRunner(mapperService, producer, decoderManager, cfg.Vin)

	if err := runner.Run(ctx, serialReader, vcan); err != nil {
		log.Printf("[ERR] Main: %v", err)

		if err != context.Canceled {
			log.Printf("[ERR] Main: pipeline failed: %v", err)
		}
	}

	log.Println("[INFO] Main: shutdown complete.")
}
