package firmata

import (
	"fmt"
)

type VersionCompatible int

const (
	VerNonCompatible VersionCompatible = iota
	VerCompatible
	VerSame
)

type PeerVersion struct {
	Major byte
	Minor byte
	Name  string
}

type Version struct {
	Client     PeerVersion
	Server     PeerVersion
	Compatible VersionCompatible
}

func NewVersion(cmajor, cminor, smajor, sminor byte) *Version {
	v := &Version{
		Client: PeerVersion{
			Major: cmajor,
			Minor: cminor,
			Name:  fmt.Sprintf("v%d.%d", cmajor, cminor),
		},
		Server: PeerVersion{
			Major: smajor,
			Minor: sminor,
			Name:  fmt.Sprintf("v%d.%d", smajor, sminor),
		},
	}
	v.Compatible = v.compatible()
	return v
}

func NewFirmwareVersion(smajor, sminor byte) *Version {
	return NewVersion(FIRMWARE_MAJOR_VERSION, FIRMWARE_MINOR_VERSION, smajor, sminor)
}

func NewProtocalVersion(smajor, sminor byte) *Version {
	return NewVersion(PROTOCOL_MAJOR_VERSION, PROTOCOL_MINOR_VERSION, smajor, sminor)
}

func (vs *Version) compatible() (c VersionCompatible) {
	if vs.Client.Major != vs.Server.Major {
		return VerNonCompatible
	}
	if vs.Client.Minor != vs.Server.Minor {
		return VerCompatible
	}
	return VerSame
}

type VersionInfo struct {
	ProtocolVersion *Version
	FirmwareVersion *Version
	FirmwareName    []byte
}
