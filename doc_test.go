package lifxlan_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"sync"
	"time"

	"go.yhsif.com/lifxlan"
)

// This example demonstrates how to do device discovery.
func Example_deviceDiscovery() {
	// Config values that should be initialized with proper args in real code.
	var (
		// The target device you want to find.
		target lifxlan.Target
		// The discover timeout.
		timeout time.Duration
	)

	// It's important to be able to cancel the context.
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	deviceChan := make(chan lifxlan.Device)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lifxlan.Discover(ctx, deviceChan, ""); err != nil {
			if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
				log.Fatalf("Discover failed: %v", err)
			}
		}
	}()

	for device := range deviceChan {
		if !device.Target().Matches(target) {
			continue
		}

		wg.Add(1)
		go func(device lifxlan.Device) {
			defer wg.Done()
			// TODO: handle device

			// If you are satisfied with the device(s) found,
			// you can cancel the context now:
			cancel()
		}(device)
	}

	wg.Wait()
}

// This example demonstrates how to send a message and wait for the ack.
//
// Please note that this example assumes that no other replies besides ack are
// expected.
// If this message will cause a response,
// you should write your own read loop instead of using WaitForAcks.
func Example_sendMessageWithAck() {
	// Should actually be a proper struct according to the Protocol definition.
	type payloadType struct{}
	// Config values that should be initialized with proper args in real code.
	var (
		// Should come with a timeout, or WaitForAcks might wait forever.
		ctx context.Context
		// The discovered device to use.
		device lifxlan.Device
		// The actual message type to be sent.
		message lifxlan.MessageType
		// The actual payload values.
		payload payloadType
	)

	conn, err := device.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	seq, err := device.Send(
		ctx,
		conn,
		lifxlan.FlagAckRequired,
		message,
		&payload, // could be nil if this message doesn't need payload.
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := lifxlan.WaitForAcks(ctx, conn, device.Source(), seq); err != nil {
		log.Fatal(err)
	}
}

// This example demonstrates how to send a message and read the response.
func Example_sendMessageWithResponse() {
	// Should actually be proper structs according to the Protocol definition.
	type (
		payloadType     struct{}
		respPayloadType struct{}
	)
	// Config values that should be initialized with proper args in real code.
	var (
		// Should come with a timeout, or we might wait forever.
		ctx context.Context
		// The discovered device to use.
		device lifxlan.Device
		// The actual message type to be sent.
		message lifxlan.MessageType
		// The actual payload values.
		payload payloadType
		// The response message type.
		respMessage lifxlan.MessageType
	)

	conn, err := device.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if ctx.Err() != nil {
		log.Fatal(ctx.Err())
	}

	seq, err := device.Send(
		ctx,
		conn,
		0, // flags, not requiring ack because this message will get a response.
		message,
		&payload, // could be nil if this message doesn't need payload.
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := lifxlan.ReadNextResponse(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		if resp.Sequence != seq || resp.Source != device.Source() {
			continue
		}
		if resp.Message != respMessage {
			continue
		}

		var raw respPayloadType
		r := bytes.NewReader(resp.Payload)
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			log.Fatal(err)
		}
		// TODO: handle payload value in raw
		return
	}
}
