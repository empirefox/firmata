package grpci

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/empirefox/firmata/pkg/dial"
	"github.com/empirefox/firmata/pkg/firmata"
	"github.com/empirefox/firmata/pkg/pb"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var empty = &emptypb.Empty{}

type temporary interface {
	Temporary() bool
}

// var _ pb.TransportServer = new(Server)

type Server struct {
	pb.UnimplementedTransportServer
	log *zerolog.Logger

	ApiVersion *pb.Version_Peer

	Boards    *pb.BoardsResponse
	BoardById map[string]*pb.Board

	Integration   *pb.Integration
	TotalFirmatas uint32

	Config      *pb.Config
	TotalGroups uint32

	// should not be locked by others
	instanceMu      sync.Mutex
	instances       []*Instance
	instanceBuilds  []bool
	instanceTmpDown []bool

	onServerMessageMu     sync.Mutex
	onSeverMessageSenders []pb.Transport_OnServerMessageServer
}

func NewServer(ctx context.Context,
	log *zerolog.Logger,
	apiVersion *pb.Version_Peer,
	boards []*pb.Board,
	integration *pb.Integration,
	config *pb.Config) *Server {
	boardById := make(map[string]*pb.Board, len(boards))
	for _, b := range boards {
		boardById[b.Id] = b
	}
	totalFirmatas := uint32(len(integration.GetFirmatas()))
	s := &Server{
		log:        log,
		ApiVersion: apiVersion,

		Boards:    &pb.BoardsResponse{Boards: boards},
		BoardById: boardById,

		Integration:   integration,
		TotalFirmatas: totalFirmatas,

		Config:      config,
		TotalGroups: uint32(len(config.GetGroups())),

		instances:       make([]*Instance, totalFirmatas),
		instanceBuilds:  make([]bool, totalFirmatas),
		instanceTmpDown: make([]bool, totalFirmatas),
	}
	go s.connectFirmatasDaemon(ctx)
	return s
}

func (s *Server) connectFirmatasDaemon(ctx context.Context) {
	t := time.Duration(s.Integration.TryConnectEverySecond) * time.Second
	s.tryConnectFirmatas(ctx)
	for {
		select {
		case <-time.After(t):
			s.tryConnectFirmatas(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) tryConnectFirmatas(ctx context.Context) {
	s.instanceMu.Lock()
	defer s.instanceMu.Unlock()
	for idx := uint32(0); idx < s.TotalFirmatas; idx++ {
		if !s.Integration.Firmatas[idx].ManualConnect &&
			s.instances[idx] == nil &&
			!s.instanceBuilds[idx] &&
			!s.instanceTmpDown[idx] {
			go s.connectFirmata(ctx, idx)
		}
	}
}

type FirmataData struct {
	Index    uint32
	PbConfig *pb.Firmata
}

func (s *Server) tryConnectFirmata(ctx context.Context, idx uint32) (err error) {
	if s.TotalFirmatas == 0 || idx >= s.TotalFirmatas {
		return fmt.Errorf("integration.firmata out of index: %d", idx)
	}

	s.instanceMu.Lock()
	defer s.instanceMu.Unlock()

	inst := s.instances[idx]
	if inst != nil {
		return
	}
	if s.instanceBuilds[idx] {
		return
	}
	s.instanceBuilds[idx] = true
	if s.instanceTmpDown[idx] {
		return
	}

	go s.connectFirmata(ctx, idx)
	return nil
}

// must be run as gorouting
func (s *Server) connectFirmata(ctx context.Context, idx uint32) (err error) {
	defer func() {
		s.instanceMu.Lock()
		s.instanceBuilds[idx] = false
		s.instanceMu.Unlock()
	}()

	var inst *Instance
	pbConfig := s.Integration.Firmatas[idx]
	firmataConfig := &firmata.Config{
		OnConnected: func(f *firmata.Firmata) {
			s.log.Debug().Str("type", "connected").
				Str("firmata", pbConfig.Name).Send()
			// init group pins
			for _, g := range s.Config.Groups {
				for _, p := range g.Pins {
					if p.FirmataIndex == idx {
						var dx byte
						switch p.Id.(type) {
						case *pb.Group_Pin_Ax:
							dx = f.AnalogPins[p.GetAx()].Dx
							p.Id = &pb.Group_Pin_Dx{Dx: uint32(dx)}
						case *pb.Group_Pin_Dx:
							dx = byte(p.GetDx())
						case *pb.Group_Pin_GpioName:
							dx = f.DxByName[p.GetGpioName()]
							p.Id = &pb.Group_Pin_Dx{Dx: uint32(dx)}
						}
						err := f.SetPinMode_l(dx, byte(p.Mode))
						if err != nil {
							s.log.Err(err).Str("firmata", pbConfig.Name).Send()
						} else {
							err = f.SetPinValue_l(dx, uint32(p.Value))
							if err != nil {
								s.log.Err(err).Str("firmata", pbConfig.Name).Send()
							}
						}
					}
				}
			}
			// init AutoHigh
			data := f.Config.Data.(*FirmataData)
			for _, w := range pbConfig.Wiring {
				if w.AutoHigh {
					var first byte
					switch w.From.First.(type) {
					case *pb.Wiring_FirmataPins_Ax:
						first = f.AnalogPins[w.From.GetAx()].Dx
					case *pb.Wiring_FirmataPins_Dx:
						first = byte(w.From.GetDx())
					case *pb.Wiring_FirmataPins_GpioName:
						first = f.DxByName[w.From.GetGpioName()]
					}

					if w.From.Slice == nil {
						err := f.SetDigitalPinHigh_l(first)
						if err != nil {
							s.log.Err(err).Str("firmata", pbConfig.Name).Send()
						}
					} else {
						var last byte
						switch w.From.Slice.(type) {
						case *pb.Wiring_FirmataPins_LastAx:
							last = f.AnalogPins[w.From.GetLastAx()].Dx
						case *pb.Wiring_FirmataPins_LastDx:
							last = byte(w.From.GetLastDx())
						case *pb.Wiring_FirmataPins_LastGpioName:
							last = f.DxByName[w.From.GetLastGpioName()]
						}
						for i := first; i <= last; i++ {
							err := f.SetDigitalPinHigh_l(i)
							if err != nil {
								s.log.Err(err).Str("firmata", pbConfig.Name).Send()
							}
						}
					}
				}
			}

			s.instanceMu.Lock()
			if s.instances[data.Index] != nil {
				s.log.Error().Msg("bugs build firmata instance!!!")
			}
			s.instances[data.Index] = inst
			s.instanceMu.Unlock()

			s.log.Debug().Str("firmata", pbConfig.Name).
				Msg("added to instannce")

			out := &pb.ServerMessage{
				Type: &pb.ServerMessage_Connected{
					Connected: inst.ToPb_l(),
				},
			}
			go s.broadcastServerMessage(out)
		},
		OnAnalogMessage: func(f *firmata.Firmata, pin *firmata.Pin) {
			data := f.Config.Data.(*FirmataData)
			out := &pb.ServerMessage{
				Type: &pb.ServerMessage_Analog_{
					Analog: &pb.ServerMessage_Analog{
						Firmata: data.Index,
						Pin:     uint32(pin.Dx),
						Value:   pin.Value_l,
					},
				},
			}
			go s.broadcastServerMessage(out)
		},
		OnDigitalMessage: func(f *firmata.Firmata, port byte, pins byte, values byte) {
			data := f.Config.Data.(*FirmataData)
			out := &pb.ServerMessage{
				Type: &pb.ServerMessage_Digital_{
					Digital: &pb.ServerMessage_Digital{
						Firmata: data.Index,
						Port:    uint32(port),
						Pins:    uint32(pins),
						Values:  uint32(values),
					},
				},
			}
			go s.broadcastServerMessage(out)
		},
		OnPinState: func(f *firmata.Firmata, pin *firmata.Pin) {
			// ignore
		},
		OnI2cReply: func(f *firmata.Firmata, reply *firmata.I2cReply) {
			// other business here
		},
		OnStringData: func(f *firmata.Firmata, b []byte) {
			data := f.Config.Data.(*FirmataData)
			s.log.Debug().Str("firmata", data.PbConfig.Name).
				Str("OnStringData", string(b)).Send()
		},
		OnSysexResponse: func(f *firmata.Firmata, buf []byte) {
			// ignore by now
		},
		Data: &FirmataData{
			Index:    idx,
			PbConfig: pbConfig,
		},
	}

	for {
		// dialing
		s.log.Debug().Str("type", "dialing").Str("firmata", pbConfig.Name).Send()
		s.broadcastConnection(idx, pb.ServerMessage_Connecting_dialing)
		c, err := dial.Dial(ctx, pbConfig.Dial)
		if err != nil {
			s.log.Debug().Str("type", "dialing").
				Str("firmata", pbConfig.Name).
				Err(err).
				Send()
			if e, ok := err.(temporary); ok && e.Temporary() {
				// dialTemporaryFail
				s.broadcastConnection(idx, pb.ServerMessage_Connecting_dialTemporaryFail)
				if s.connectRetry(ctx, idx) {
					continue
				}

				// disconnected
				s.broadcastConnection(idx, pb.ServerMessage_Connecting_disconnected)
				return err
			}
			// dialFatalError
			s.broadcastConnection(idx, pb.ServerMessage_Connecting_dialFatalError)
			return err
		}

		inst = &Instance{
			log:     s.log,
			index:   idx,
			config:  pbConfig,
			firmata: firmata.NewFirmata(c, firmataConfig),
		}

		err = inst.Handshake(ctx)
		if err != nil {
			// handshakeError
			s.log.Debug().Str("type", "handshake").
				Str("firmata", pbConfig.Name).
				Str("err", err.Error()).
				Err(inst.firmata.ClosedError_l).
				Send()
			s.broadcastConnection(idx, pb.ServerMessage_Connecting_handshakeError)
			if s.connectRetry(ctx, idx) {
				continue
			}

			// disconnected
			s.broadcastConnection(idx, pb.ServerMessage_Connecting_disconnected)
			return err
		}

		// ok
		go s.waitFirmataClosed(inst)
		return nil
	}
}

func (s *Server) connectRetry(ctx context.Context, idx uint32) (retry bool) {
	s.instanceMu.Lock()
	tmpDown := s.instanceTmpDown[idx]
	s.instanceMu.Unlock()
	if tmpDown {
		return false
	}

	sleep := s.Integration.Firmatas[idx].ConnectRetrySecond

	select {
	case <-time.After(time.Second * time.Duration(sleep)):
	case <-ctx.Done():
		return false
	}

	s.instanceMu.Lock()
	tmpDown = s.instanceTmpDown[idx]
	s.instanceMu.Unlock()
	return !tmpDown
}

func (s *Server) disconnectFirmata(ctx context.Context, idx uint32) {
	s.instanceMu.Lock()
	inst := s.instances[idx]
	s.instanceMu.Unlock()
	if inst != nil {
		inst.firmata.Close()
	}
}

func (s *Server) waitFirmataClosed(inst *Instance) {
	<-inst.firmata.CloseNotify()
	s.instanceMu.Lock()
	s.instances[inst.index] = nil
	s.instanceMu.Unlock()
	s.log.Debug().Str("firmata", inst.config.Name).
		Msg("removed from instannce")
	s.broadcastConnection(inst.index, pb.ServerMessage_Connecting_disconnected)
}

func (s *Server) loopFromFirmata(firmataIndex uint32, fn func(*Instance) error) error {
	if s.TotalFirmatas == 0 || firmataIndex >= s.TotalFirmatas {
		return fmt.Errorf("config.firmatas out of index: %d", firmataIndex)
	}

	s.instanceMu.Lock()
	inst := s.instances[firmataIndex]
	s.instanceMu.Unlock()
	if inst == nil {
		return fmt.Errorf("firmata disconnected")
	}

	return inst.firmata.WaitLoop(func() error {
		return fn(inst)
	})
}

func (s *Server) loopFromGroup(group uint32, gpin uint32, pre, fn func(*Instance, *pb.Group_Pin) error) (err error) {
	if s.TotalGroups == 0 || group >= s.TotalGroups {
		return fmt.Errorf("config.groups out of index: %d", group)
	}

	g := s.Config.Groups[group]
	gpinSize := uint32(len(g.Pins))
	if gpinSize == 0 || gpin >= gpinSize {
		return fmt.Errorf("config.groups[%d].pins out of index: %d", group, gpin)
	}
	gp := g.Pins[gpin]

	s.instanceMu.Lock()
	inst := s.instances[gp.FirmataIndex]
	s.instanceMu.Unlock()
	if inst == nil {
		return fmt.Errorf("firmata disconnected")
	}

	if pre != nil {
		err = pre(inst, gp)
		if err != nil {
			return
		}
	}
	err = inst.firmata.WaitLoop(func() error {
		return fn(inst, gp)
	})
	return err
}

func (s *Server) sendInstancesTo(sender pb.Transport_OnServerMessageServer) (err error) {
	s.instanceMu.Lock()
	defer s.instanceMu.Unlock()
	for i, inst := range s.instances {
		var out *pb.ServerMessage
		if inst == nil {
			out = &pb.ServerMessage{
				Type: &pb.ServerMessage_Connecting_{
					Connecting: &pb.ServerMessage_Connecting{
						Firmata: uint32(i),
						Status:  pb.ServerMessage_Connecting_disconnected,
					},
				},
			}
		} else {
			pbInst, e := inst.ToPb()
			if e != nil {
				// disconnected now, ignore
				continue
			}
			out = &pb.ServerMessage{
				Type: &pb.ServerMessage_Connected{
					Connected: pbInst,
				},
			}
		}
		err = sender.Send(out)
		if err != nil {
			return
		}
	}
	return
}

func (s *Server) broadcastConnection(idx uint32, status pb.ServerMessage_Connecting_Status) {
	out := &pb.ServerMessage{
		Type: &pb.ServerMessage_Connecting_{
			Connecting: &pb.ServerMessage_Connecting{
				Firmata: idx,
				Status:  status,
			},
		},
	}
	s.broadcastServerMessage(out)
}

func (s *Server) broadcastServerMessage(out *pb.ServerMessage) {
	s.onServerMessageMu.Lock()
	defer s.onServerMessageMu.Unlock()
	for _, sender := range s.onSeverMessageSenders {
		sender.Send(out)
	}
}

func (s *Server) GetApiVersion(ctx context.Context, in *emptypb.Empty) (*pb.Version_Peer, error) {
	return s.ApiVersion, nil
}
func (s *Server) ListBoards(ctx context.Context, in *emptypb.Empty) (*pb.BoardsResponse, error) {
	return s.Boards, nil
}
func (s *Server) GetBoard(ctx context.Context, in *wrapperspb.StringValue) (*pb.Board, error) {
	return s.BoardById[in.GetValue()], nil
}
func (s *Server) GetIntegration(ctx context.Context, in *emptypb.Empty) (*pb.Integration, error) {
	return s.Integration, nil
}
func (s *Server) GetConfig(ctx context.Context, in *emptypb.Empty) (*pb.Config, error) {
	return s.Config, nil
}

func (s *Server) OnServerMessage(in *emptypb.Empty, stream pb.Transport_OnServerMessageServer) error {
	s.onServerMessageMu.Lock()
	s.onSeverMessageSenders = append(s.onSeverMessageSenders, stream)
	s.onServerMessageMu.Unlock()

	s.sendInstancesTo(stream)
	<-stream.Context().Done()

	s.onServerMessageMu.Lock()
	for i, sender := range s.onSeverMessageSenders {
		if sender == stream {
			last := len(s.onSeverMessageSenders) - 1
			s.onSeverMessageSenders[i] = s.onSeverMessageSenders[last]
			s.onSeverMessageSenders = s.onSeverMessageSenders[:last]
			break
		}
	}
	s.onServerMessageMu.Unlock()
	return nil
}

func (s *Server) Connect(ctx context.Context, in *pb.FirmataIndex) (*emptypb.Empty, error) {
	s.instanceMu.Lock()
	s.instanceTmpDown[in.Firmata] = false
	s.instanceMu.Unlock()
	return empty, s.tryConnectFirmata(ctx, in.Firmata)
}
func (s *Server) Disconnect(ctx context.Context, in *pb.FirmataIndex) (*emptypb.Empty, error) {
	s.instanceMu.Lock()
	s.instanceTmpDown[in.Firmata] = true
	s.instanceMu.Unlock()
	s.disconnectFirmata(ctx, in.Firmata)
	return empty, nil
}
func (s *Server) SetPinMode(ctx context.Context, in *pb.SetPinModeRequest) (*emptypb.Empty, error) {
	s.instanceMu.Lock()
	inst := s.instances[in.Firmata]
	s.instanceMu.Unlock()
	if inst == nil {
		return nil, fmt.Errorf("firmata disconnected")
	}

	f := inst.firmata
	err := f.WaitLoop(func() error {
		return f.SetPinMode_l(byte(in.Dx), byte(in.Mode))
	})
	// TODO broadcast?
	return empty, err
}
func (s *Server) TriggerDigitalPin(ctx context.Context, in *pb.TriggerDigitalPinRequest) (*emptypb.Empty, error) {
	var instance *Instance
	var f *firmata.Firmata
	var dx byte
	var triggerMs uint32
	var values1 byte = 1
	var values2 byte = 0
	err := s.loopFromGroup(in.Group, in.Gpin,
		func(inst *Instance, gp *pb.Group_Pin) error {
			instance = inst
			f = inst.firmata
			dx = byte(gp.GetDx())

			var isButton bool = true
			var swtch *pb.Group_Switch
			btn := gp.GetButton()
			if btn == nil {
				swtch = gp.GetSwitch()
				if swtch == nil {
					return fmt.Errorf("TriggerDigitalPin accepts button/switch type: %t",
						gp.Type)
				}
				isButton = false
			}

			var lowLevelTrigger bool
			if isButton {
				lowLevelTrigger = btn.LowLevelTrigger
				triggerMs = btn.TriggerMs
			} else {
				lowLevelTrigger = swtch.LowLevelTrigger
				triggerMs = swtch.TriggerMs
			}

			if lowLevelTrigger {
				values1 = 0
				values2 = 1
			}

			if in.RealtimeTriggerMs != 0 {
				triggerMs = in.RealtimeTriggerMs
			} else if triggerMs == 0 {
				return fmt.Errorf("TriggerDigitalPinRequest.RealtimeTriggerMs is required")
			}
			return nil
		},
		func(inst *Instance, gp *pb.Group_Pin) error {
			s.log.Debug().Str("firmata", instance.config.Name).
				Uint8("dx", dx).Uint8("v", values1).Send()
			return f.SetDigitalPinValue_l(dx, values1)
		})
	if err != nil {
		return nil, err
	}

	data := pb.ServerMessage_Digital{
		Firmata: instance.index,
		Port:    uint32(dx / 8),
		Pins:    1 << (dx % 8),
		Values:  uint32(values2) << (dx % 8),
	}
	out := &pb.ServerMessage{
		Type: &pb.ServerMessage_Digital_{
			Digital: &data,
		},
	}

	s.broadcastServerMessage(out)

	time.Sleep(time.Duration(triggerMs) * time.Millisecond)
	err = f.WaitLoop(func() error {
		s.log.Debug().Str("firmata", instance.config.Name).
			Uint8("dx", dx).Uint8("v", values2).Send()
		return f.SetDigitalPinValue_l(dx, values2)
	})

	data.Values = uint32(values2) << (dx % 8)
	s.broadcastServerMessage(out)

	return empty, err
}
func (s *Server) SetPinValue(ctx context.Context, in *pb.SetPinValueRequest) (*emptypb.Empty, error) {
	var instance *Instance
	var dx byte
	err := s.loopFromGroup(in.Group, in.Gpin, nil, func(inst *Instance, gp *pb.Group_Pin) error {
		instance = inst
		dx = byte(gp.GetDx())
		s.log.Debug().Str("firmata", inst.config.Name).
			Uint8("dx", dx).Uint32("v", in.Value).Send()
		return inst.firmata.SetPinValue_l(dx, in.Value)
	})
	if err != nil {
		return nil, err
	}

	var out *pb.ServerMessage
	if instance.firmata.Pins[dx].IsAnalog() {
		out = &pb.ServerMessage{
			Type: &pb.ServerMessage_Analog_{
				Analog: &pb.ServerMessage_Analog{
					Firmata: instance.index,
					Pin:     uint32(dx),
					Value:   in.Value,
				},
			},
		}
	} else {
		out = &pb.ServerMessage{
			Type: &pb.ServerMessage_Digital_{
				Digital: &pb.ServerMessage_Digital{
					Firmata: instance.index,
					Port:    uint32(dx / 8),
					Pins:    1 << (dx % 8),
					Values:  in.Value << (dx % 8),
				},
			},
		}
	}

	s.broadcastServerMessage(out)
	return empty, nil
}
func (s *Server) ReportDigital(ctx context.Context, in *pb.ReportDigitalRequest) (*emptypb.Empty, error) {
	err := s.loopFromFirmata(in.Firmata, func(inst *Instance) error {
		return inst.firmata.ReportDigital_l(byte(in.Port), in.Enable)
	})
	return empty, err
}
func (s *Server) ReportAnalog(ctx context.Context, in *pb.ReportAnalogRequest) (*emptypb.Empty, error) {
	err := s.loopFromFirmata(in.Firmata, func(inst *Instance) error {
		return inst.firmata.ReportAnalog_l(byte(in.Pin), in.Enable)
	})
	return empty, err
}
func (s *Server) WriteString(ctx context.Context, in *pb.WriteStringRequest) (*emptypb.Empty, error) {
	err := s.loopFromFirmata(in.Firmata, func(inst *Instance) error {
		return inst.firmata.StringWrite_l([]byte(in.Data))
	})
	return empty, err
}
func (s *Server) SetSamplingInterval(ctx context.Context, in *pb.SetSamplingIntervalRequest) (*emptypb.Empty, error) {
	err := s.loopFromFirmata(in.Firmata, func(inst *Instance) error {
		return inst.firmata.SamplingInterval_l(in.Ms)
	})
	return empty, err
}
