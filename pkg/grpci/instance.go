package grpci

import (
	"context"

	"github.com/empirefox/firmata/pkg/dial"
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

func NewInstance(log *zerolog.Logger, index uint32,
	pbConfig *pb.Firmata, firmataConfig *firmata.Config) (*Instance, error) {
	c, err := dial.Dial(pbConfig.Dial)
	if err != nil {
		return nil, err
	}

	return &Instance{
		log:     log,
		index:   index,
		config:  pbConfig,
		firmata: firmata.NewFirmata(c, firmataConfig),
	}, nil
}

func (inst *Instance) Handshake(ctx context.Context) error {
	return inst.firmata.Handshake(ctx)
}

func (inst *Instance) Data() *FirmataData {
	return inst.firmata.Config.Data.(*FirmataData)
}

func (inst *Instance) ToPb() (out *pb.Instance, err error) {
	f := inst.firmata
	err = f.WaitLoop(func() error {
		out = &pb.Instance{
			Firmata:          inst.config.Name,
			FirmataIndex:     inst.index,
			ProtocolVersion:  f.ProtocolVersion,
			FirmwareVersion:  f.FirmwareVersion,
			Pins:             f.PinsToPb_l(),
			PortConfigInputs: f.PortConfigInputs_l[:f.TotalPorts],
		}
		return nil
	})
	return
}