package grpci

import (
	"context"
	"time"

	"github.com/empirefox/firmata/pkg/firmata"
	"github.com/empirefox/firmata/pkg/pb"
	"github.com/rs/zerolog"
)

type Instance struct {
	log *zerolog.Logger

	index   uint32
	config  *pb.Firmata
	firmata *firmata.Firmata
}

func (inst *Instance) Handshake(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return inst.firmata.Handshake(ctx)
}

func (inst *Instance) Data() *FirmataData {
	return inst.firmata.Config.Data.(*FirmataData)
}

func (inst *Instance) ToPb_l() (out *pb.Instance) {
	f := inst.firmata
	return &pb.Instance{
		Firmata:          inst.config.Name,
		FirmataIndex:     inst.index,
		ProtocolVersion:  f.ProtocolVersion,
		FirmwareVersion:  f.FirmwareVersion,
		Pins:             f.PinsToPb_l(),
		PortConfigInputs: f.PortConfigInputs_l[:f.TotalPorts],
	}
}

func (inst *Instance) ToPb() (out *pb.Instance, err error) {
	f := inst.firmata
	err = f.WaitLoop(func() error {
		out = inst.ToPb_l()
		return nil
	})
	return
}
