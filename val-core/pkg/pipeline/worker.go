package pipeline

import (
	"context"
	"net"
	"strings"
	"time"

	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"
	pb "ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/api"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/internal/util"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/decoding"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/kafka"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/mapper"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/nmea"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/source"
)

func ParseTelemetryLoop(ctx context.Context, in <-chan *models.DecodedSignal, out chan<- *pb.TelematicsBatch, mp *mapper.Service, vin string) {
	const batchSize = 10
	buffer := make([]*models.DecodedSignal, 0, batchSize)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		batchMsg, err := MarshalSignal(vin, mp, buffer)
		if err != nil {
			log.Warn("error marshalling batch", "message", err)
		} else {
			select {
			case out <- batchMsg:
			case <-ctx.Done():
				return

			}
		}

		buffer = buffer[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case decodedFrame, ok := <-in:
			if !ok {
				flush()
				return
			}

			buffer = append(buffer, decodedFrame)

			if len(buffer) >= batchSize {
				flush()
				ticker.Reset(10 * time.Second)
			}
		case <-ticker.C:
			flush()
		}
	}
}

func PublishTelemetryLoop(ctx context.Context, producer *kafka.Producer, in <-chan *pb.TelematicsBatch) {
	var topic string

	for {
		select {
		case <-ctx.Done():
			log.Info("context cancelled, flushing producer")
			producer.Flush()
			log.Info("producer flushed. exiting loop")
			return
		case parsedMsg, ok := <-in:
			if !ok {
				log.Info("channel closed, flushing producer")
				producer.Flush()
				log.Info("producer flushed. exiting loop")
				return
			}

			topic = util.If(parsedMsg.GetVin() == "f1", "game-telematics", "vehicle-telematics")

			if err := producer.PublishTelematics(parsedMsg, &topic); err != nil {
				log.Warn("kafka publish error", "message", err)
			}
		}
	}
}

func ReadUDPLoop(ctx context.Context, u *source.UDPConnection, out chan<- []byte) {
	defer u.Close()

	buf := make([]byte, 2048)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, _, err := u.Conn.ReadFromUDP(buf)
			if err != nil {
				// Timeout is expected, other errors are not
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Warn("udp read error", "message", err)

				continue
			}

			packet := make([]byte, n)
			copy(packet, buf[:n])

			out <- packet
		}
	}
}

func DecodeUDPLoop(ctx context.Context, d *decoding.Manager, in <-chan []byte, out chan<- *models.DecodedSignal) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-in:
			if !ok {
				return
			}

			log.Info("processing game packet")

			signals, err := d.DecodeGame(data)
			if err != nil {
				log.Warn("unable to decode game packet", "message", err)
				continue
			}

			for _, sig := range signals {
				log.Info("decoded packet")

				select {
				case out <- sig:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func PublishPacketLoop(ctx context.Context, producer *kafka.Producer, in <-chan *pb.TelematicsBatch) {
	topic := "game-telematics"

	for {
		select {
		case <-ctx.Done():
			log.Info("context cancelled, flushing producer")
			producer.Flush()
			log.Info("producer flushed, exiting loop")
			return
		case parsedMsg, ok := <-in:
			if !ok {
				log.Info("channel closed, flushing producer")
				producer.Flush()
				log.Info("producer flushed, exiting loop")
				return
			}

			if err := producer.PublishTelematics(parsedMsg, &topic); err != nil {
				log.Warn("kafka publish error", "message", err)
			}
		}
	}
}

func ReadCanFrameLoop(ctx context.Context, sr *source.VcanConnection, out chan<- can.Frame) {
	type readResult struct {
		frame can.Frame
		err   error
	}
	resultCh := make(chan readResult)

	go func() {
		defer close(resultCh)

		for sr.Recv.Receive() {
			frame := sr.Recv.Frame()

			log.Info("received frame")

			select {
			case resultCh <- readResult{frame, nil}:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-resultCh:
			if !ok {
				return
			}

			out <- frame.frame

		}
	}
}

func DecodeFrameLoop(ctx context.Context, d *decoding.Manager, in <-chan can.Frame, out chan<- *models.DecodedSignal) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-in:
			if !ok {
				return
			}

			log.Info("processing frame", "frame_id", frame.ID)

			signals, err := d.Decode(frame)
			if err != nil {
				log.Warn("frame decoding error", "message", err, "frame_id", frame.ID)
				continue
			}

			for _, sig := range signals {
				log.Info("decoded signal")
				select {
				case out <- sig:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func ReadSerialLoop(ctx context.Context, sr *source.SerialReader, out chan<- string) {
	if sr == nil {
		log.Warn("serial reader is nil, skipping")
		return
	}

	type readResult struct {
		line string
		err  error
	}
	resultCh := make(chan readResult)

	go func() {
		defer close(resultCh)

		var buffer string

		for {
			line, err := sr.ReadLine()
			if err != nil {
				resultCh <- readResult{"", err}
				continue
			}

			buffer += line

			if strings.Contains(buffer, "\n") {
				parts := strings.Split(buffer, "\n")

				for i := 0; i < len(parts)-1; i++ {
					s := strings.TrimSpace(parts[i])
					if s != "" {
						select {
						case resultCh <- readResult{s, nil}:
						case <-ctx.Done():
							return
						}
					}
				}

				buffer = parts[len(parts)-1]
			}
		}

	}()

	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-resultCh:
			if !ok {
				return
			}

			if res.err != nil {
				log.Warn("serial read error", "message", res.err)
			}

			out <- res.line
		}
	}
}

func ParseLocationLoop(ctx context.Context, in <-chan string, out chan<- *pb.LocationRequest, vin string) {
	for {
		select {
		case <-ctx.Done():
			return
		case sentence, ok := <-in:
			if !ok {
				return
			}

			s, err := nmea.ParseSentence(sentence)
			if err != nil {
				log.Warn("nmea parse error", "message", err)
				continue
			}

			loc, err := nmea.SentenceToLocation(s, vin)
			if err != nil {
				continue
			}
			out <- loc
		}
	}
}

func PublishLocationLoop(ctx context.Context, producer *kafka.Producer, in <-chan *pb.LocationRequest) {
	topic := "vehicle-location"

	for {
		select {
		case <-ctx.Done():
			return
		case loc, ok := <-in:
			if !ok {
				return
			}

			if err := producer.PublishLocation(loc, &topic); err != nil {
				log.Warn("kafka publish error", "message", err)
			}
		}
	}
}
