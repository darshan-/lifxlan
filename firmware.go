package lifxlan

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
)

// RawStateHostFirmwarePayload defines the struct to be used for encoding and
// decoding.
//
// https://lan.developer.lifx.com/v2.0/docs/device-messages#section-stateversion-33
type RawStateHostFirmwarePayload struct {
	_            uint64 // build
	_            uint64 // reserved
	VersionMinor uint16
	VersionMajor uint16
}

// ToFirmware converts RawStateHostFirmwarePayload into FirmwareUpgrade
// with empty Features.
func (raw RawStateHostFirmwarePayload) ToFirmware() FirmwareUpgrade {
	return FirmwareUpgrade{
		Major: raw.VersionMajor,
		Minor: raw.VersionMinor,
	}
}

func (d *device) Firmware() *FirmwareUpgrade {
	return &d.firmware
}

func (d *device) GetFirmware(ctx context.Context, conn net.Conn) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	if conn == nil {
		newConn, err := d.Dial()
		if err != nil {
			return err
		}
		defer newConn.Close()
		conn = newConn

		select {
		default:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	seq, err := d.Send(
		ctx,
		conn,
		0, // flags
		GetHostFirmware,
		nil, // payload
	)
	if err != nil {
		return err
	}

	for {
		resp, err := ReadNextResponse(ctx, conn)
		if err != nil {
			return err
		}
		if resp.Sequence != seq || resp.Source != d.Source() {
			continue
		}
		if resp.Message != StateHostFirmware {
			continue
		}

		var raw RawStateHostFirmwarePayload
		r := bytes.NewReader(resp.Payload)
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			return err
		}

		d.firmware = raw.ToFirmware()
		return nil
	}
}
