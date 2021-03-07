package lifxlan

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

func (d *device) Send(
	ctx context.Context,
	conn net.Conn,
	flags AckResFlag,
	message MessageType,
	payload interface{},
) (seq uint8, err error) {
	if ctx.Err() != nil {
		err = ctx.Err()
		return
	}

	var msg []byte
	seq = d.NextSequence()
	buf := new(bytes.Buffer)
	if payload != nil {
		if err = binary.Write(buf, binary.LittleEndian, payload); err != nil {
			return
		}
	}
	msg, err = GenerateMessage(
		NotTagged,
		d.Source(),
		d.Target(),
		flags,
		seq,
		message,
		buf.Bytes(),
	)
	if err != nil {
		return
	}

	if ctx.Err() != nil {
		err = ctx.Err()
		return
	}

	var n int
	n, err = conn.Write(msg)
	if err != nil {
		return
	}
	if n < len(msg) {
		err = fmt.Errorf(
			"lifxlan.Device.Send: only wrote %d out of %d bytes",
			n,
			len(msg),
		)
		return
	}

	if ctx.Err() != nil {
		err = ctx.Err()
		return
	}

	return
}
