package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/foxboron/goefi/efi/util"
)

// Subtypes of Media Device
// Section 10.3.5 - Media Device Path
const (
	_ DevicePathSubType = iota
	HardDriveDevicePath
	CDRomDevicePath
	VendorMediaDevicePath
	FilePathDevicePath
	MediaProtocolDevicePath
	PIWGFirmwareDevicePath
)

type HardDriveMediaDevicePath struct {
	EFIDevicePath
	PartitionNumber    uint32
	PartitionStart     [8]byte
	PartitionSize      [8]byte
	PartitionSignature [16]byte
	PartitionFormat    uint8
	SignatureType      uint8
}

func (h HardDriveMediaDevicePath) Format() string {
	format := []string{"MBR", "GPT"}
	if h.PartitionNumber == 0 {
		return fmt.Sprintf("HD(%d,%s,%x)",
			h.PartitionNumber,
			format[h.PartitionFormat-1],
			h.PartitionSignature)
	}
	return fmt.Sprintf("HD(%d,%s,%s,0x%x,0x%x)",
		h.PartitionNumber,
		format[h.PartitionFormat-1],
		util.BytesToGUID(h.PartitionSignature[:]).Format(),
		binary.LittleEndian.Uint64(h.PartitionStart[:]),
		binary.LittleEndian.Uint64(h.PartitionSize[:]))
}

type FileTypeMediaDevicePath struct {
	EFIDevicePath
	PathName []byte
}

func (f FileTypeMediaDevicePath) Format() string {
	return fmt.Sprintf("File(%s)", f.PathName)
}

type FirmwareFielMediaDevicePath struct {
	EFIDevicePath
	FirmwareFileName [16]byte
}

func ParseMediaDevicePath(f *bytes.Reader, efi *EFIDevicePath) EFIDevicePaths {
	switch efi.SubType {
	case HardDriveDevicePath:
		m := HardDriveMediaDevicePath{EFIDevicePath: *efi}
		// var m HardDriveMediaDevicePath
		if err := binary.Read(f, binary.LittleEndian, &m.PartitionNumber); err != nil {
			log.Fatal(err)
		}
		if err := binary.Read(f, binary.LittleEndian, &m.PartitionStart); err != nil {
			log.Fatal(err)
		}
		if err := binary.Read(f, binary.LittleEndian, &m.PartitionSize); err != nil {
			log.Fatal(err)
		}
		if err := binary.Read(f, binary.BigEndian, &m.PartitionSignature); err != nil {
			log.Fatal(err)
		}
		if err := binary.Read(f, binary.LittleEndian, &m.PartitionFormat); err != nil {
			log.Fatal(err)
		}
		if err := binary.Read(f, binary.LittleEndian, &m.SignatureType); err != nil {
			log.Fatal(err)
		}
		return m
	case FilePathDevicePath:
		file := FileTypeMediaDevicePath{EFIDevicePath: *efi}
		file.PathName = util.ReadNullString(f)
		return file
	case PIWGFirmwareDevicePath:
		file := FirmwareFielMediaDevicePath{EFIDevicePath: *efi}
		if err := binary.Read(f, binary.LittleEndian, &file.FirmwareFileName); err != nil {
			log.Fatal(err)
		}
		return file
	default:
		log.Printf("Not implemented MediaDevicePath type: %x", efi.SubType)
	}
	return nil
}
