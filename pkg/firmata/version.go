package firmata

import (
	"github.com/empirefox/firmata/pkg/pb"
)

const (
	UnkownVersionPart   = -1
	FirmataProtocolName = "github.com/firmata/protocol"
)

type Compatible = pb.Version_Compatible

type Version = pb.Version

func NewVersion(client, server *pb.Version_Peer) *Version {
	v := &Version{
		Client: client,
		Server: server,
	}
	v.Compatible = compatible(v)
	return v
}

func NewFirmwareVersion(name string, major, minor byte) *Version {
	return NewVersion(
		&pb.Version_Peer{
			Major:  int32(FIRMATA_FIRMWARE_MAJOR_VERSION),
			Minor:  int32(FIRMATA_FIRMWARE_MINOR_VERSION),
			Bugfix: int32(FIRMATA_FIRMWARE_BUGFIX_VERSION),
		},
		&pb.Version_Peer{
			Name:   name,
			Major:  int32(major),
			Minor:  int32(minor),
			Bugfix: UnkownVersionPart,
		},
	)
}

func NewProtocalVersion(name string, major, minor byte) *Version {
	return NewVersion(
		&pb.Version_Peer{
			Major:  int32(FIRMATA_PROTOCOL_MAJOR_VERSION),
			Minor:  int32(FIRMATA_PROTOCOL_MINOR_VERSION),
			Bugfix: int32(FIRMATA_PROTOCOL_BUGFIX_VERSION),
		},
		&pb.Version_Peer{
			Name:   name,
			Major:  int32(major),
			Minor:  int32(minor),
			Bugfix: UnkownVersionPart,
		},
	)
}

func compatible(v *Version) Compatible {
	if v.Client.Major != v.Server.Major {
		return pb.Version_no
	}
	if v.Client.Minor != v.Server.Minor {
		return pb.Version_yes
	}
	return pb.Version_same
}

type VersionInfo struct {
	ProtocolVersion *Version
	FirmwareVersion *Version
}
