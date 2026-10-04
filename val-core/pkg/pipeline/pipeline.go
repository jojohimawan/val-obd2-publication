package pipeline

import (
	"context"
	"log/slog"
	"sync"

	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"
	pb "ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/api"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/config"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/decoding"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/kafka"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/mapper"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/source"
)

var log *slog.Logger

type RunnerConfig struct {
	Mapper   *mapper.Service
	Producer *kafka.Producer
	Decoder  *decoding.Manager
	Vin      string
}

type Runner struct {
	mapper   *mapper.Service
	producer *kafka.Producer
	decoder  *decoding.Manager
	vin      string
}

func NewRunner(cfg RunnerConfig) *Runner {
	log = config.ComponentLogger("pipeline")

	return &Runner{
		mapper:   cfg.Mapper,
		producer: cfg.Producer,
		decoder:  cfg.Decoder,
		vin:      cfg.Vin,
	}
}

func (r *Runner) Run(
	ctx context.Context,
	sr *source.SerialReader,
	sv *source.VcanConnection,
) error {
	rawSentences := make(chan string, 50)
	locations := make(chan *pb.LocationRequest, 50)

	rawFrame := make(chan can.Frame, 50)
	decodedTelemetry := make(chan *models.DecodedSignal, 50)
	parsedTelemetry := make(chan *pb.TelematicsBatch, 50)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ReadCanFrameLoop(ctx, sv, rawFrame)
		close(rawFrame)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		DecodeFrameLoop(ctx, r.decoder, rawFrame, decodedTelemetry)
		close(decodedTelemetry)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ParseTelemetryLoop(ctx, decodedTelemetry, parsedTelemetry, r.mapper, r.vin)
		close(parsedTelemetry)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		PublishTelemetryLoop(ctx, r.producer, parsedTelemetry)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ReadSerialLoop(ctx, sr, rawSentences)
		close(rawSentences)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ParseLocationLoop(ctx, rawSentences, locations, r.vin)
		close(locations)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		PublishLocationLoop(ctx, r.producer, locations)
	}()

	wg.Wait()
	log.Info("Pipeline terminated cleanly.")
	return nil
}

func (r *Runner) RunGame(
	ctx context.Context,
	uc *source.UDPConnection,
) error {
	rawPacket := make(chan []byte)

	decodedTelemetry := make(chan *models.DecodedSignal, 50)
	parsedTelemetry := make(chan *pb.TelematicsBatch, 50)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ReadUDPLoop(ctx, uc, rawPacket)
		close(rawPacket)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		DecodeUDPLoop(ctx, r.decoder, rawPacket, decodedTelemetry)
		close(decodedTelemetry)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ParseTelemetryLoop(ctx, decodedTelemetry, parsedTelemetry, r.mapper, r.vin)
		close(parsedTelemetry)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		PublishPacketLoop(ctx, r.producer, parsedTelemetry)
	}()

	wg.Wait()
	log.Info("Pipeline terminated cleanly.")
	return nil
}
