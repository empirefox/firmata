package pbload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/empirefox/firmata/pkg/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var json = protojson.UnmarshalOptions{DiscardUnknown: true}

func LoadApiVersion() *pb.Version_Peer {
	return &pb.Version_Peer{Major: 0, Minor: 0, Bugfix: 1}
}

func LoadBoards(d string) ([]*pb.Board, error) {
	files, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}

	boards := make([]*pb.Board, 0, len(files))
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".json") {
			var board pb.Board
			err = JsonFileToPb(filepath.Join(d, f.Name()), &board)
			if err != nil {
				return nil, err
			}

			boards = append(boards, &board)
		}
	}

	return boards, nil
}

func LoadIntegration(p string) (*pb.Integration, error) {
	var v pb.Integration
	err := JsonFileToPb(p, &v)
	return &v, err
}

func LoadConfig(p string) (*pb.Config, error) {
	var v pb.Config
	err := JsonFileToPb(p, &v)
	return &v, err
}

func JsonFileToPb(p string, m protoreflect.ProtoMessage) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, m)
}

func CheckError(boards []*pb.Board, integration *pb.Integration, config *pb.Config) error {
	boardByName := make(map[string]uint32, len(boards))
	for i, b := range boards {
		if _, ok := boardByName[b.Id]; ok {
			return fmt.Errorf("duplicated board id: %s", b.Id)
		}
		boardByName[b.Id] = uint32(i)
	}

	if integration.TryConnectEverySecond == 0 {
		integration.TryConnectEverySecond = 10
	}

	firmataByName := make(map[string]uint32, len(integration.Firmatas))
	for i, f := range integration.Firmatas {
		if _, ok := firmataByName[f.Name]; ok {
			return fmt.Errorf("duplicated firmata name: %s", f.Name)
		}
		firmataByName[f.Name] = uint32(i)
	}

	for i, t := range integration.Firmatas {
		if t.ConnectRetrySecond == 0 {
			t.ConnectRetrySecond = 10
		}
		for _, w := range t.Wiring {
			w.From.FirmataIndex = uint32(i)
			if f := w.To.GetFirmata(); f != nil {
				index, ok := firmataByName[f.Firmata]
				if !ok {
					return fmt.Errorf("wire firmata not found: %s", f.Firmata)
				}
				f.FirmataIndex = index
			}
		}
	}

	for _, g := range config.Groups {
		for _, p := range g.Pins {
			index, ok := firmataByName[p.Firmata]
			if !ok {
				return fmt.Errorf("group pin of firmata not found: %s", p.Firmata)
			}
			p.FirmataIndex = index

			if dt := p.GetSwitch().GetDetect(); dt != nil {
				index, ok := firmataByName[dt.Firmata]
				if !ok {
					return fmt.Errorf("group pin of firmata not found: %s", dt.Firmata)
				}
				dt.FirmataIndex = index
			}
		}
	}

	return nil
}
