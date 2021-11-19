package pbload

import (
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
