package pipeline

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	pb "ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/api"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/mapper"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"
)

func MarshalSignal(vin string, mapper *mapper.Service, signals []*models.DecodedSignal) (*pb.TelematicsBatch, error) {
	var protoSignals []*pb.Telematics

	for _, s := range signals {
		vssPoint, found := mapper.Translate(s)
		if !found {
			log.Warn("unknown signal, skipping.")
			continue
		}

		protoSignals = append(protoSignals, &pb.Telematics{
			Source: s.Source,
			Param:  vssPoint.Path,
			Value:  &pb.Telematics_DoubleVal{DoubleVal: s.Value.(float64)},
			Unit:   vssPoint.Unit,
		})
	}

	return &pb.TelematicsBatch{
		Vin:         vin,
		CaptureTime: timestamppb.New(time.Now()),
		Signals:     protoSignals,
	}, nil
}
