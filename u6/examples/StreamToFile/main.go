package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gousb"
	"github.com/soypat/labjack/u6"
)

func main() {

	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()
	ctx.Debug(1)

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	fmt.Println(dev.DeviceDesc())

	// These values are only demonstative. not tested
	cfg := u6.StreamConfig{
		ResolutionIndex:  1,
		SamplesPerPacket: 25,
		SettlingFactor:   0,
		ScanFrequency:    1,
		ScanConfig: &u6.ScanConfig{
			ClockSpeed:   u6.ClockSpeed4Mhz,
			DivideBy256:  u6.ClockDivisionOff,
			ScanInterval: 0},
		Channels: []u6.ChannelConfig{
			{PositiveChannel: 1, GainIndex: u6.GainIndex1000, Differential: u6.DifferentialInputDisabled}, // one channel
		},
	}
	stream, err := dev.NewStream(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := stream.Start()
	if err != nil {
		log.Fatal(err)
	}

	fh, _ := os.Create("out.csv")
	defer fh.Close()

	timeout := time.After(time.Second * 30)
OUTER:
	for {
		select {
		case resp := <-ch:
			// fmt.Println("Packet: ", resp.PacketNumber)
			for _, channel := range resp.Data {
				voltage, err := channel.GetCalibratedAIN()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fh.WriteString(fmt.Sprintf("%s,%d,%d,%d,%0.8f\n", time.Now().Format(time.RFC3339Nano), channel.ChannelIndex, channel.ScanNumber, resp.PacketNumber, voltage))
				fmt.Printf("Packet=%d; ChannelIndex=%d; ScanNumber=%d; Voltage=%0.6f\n", resp.PacketNumber, channel.ChannelIndex, channel.ScanNumber, voltage)
			}

			// fmt.Println(time.Now(), voltage)
		case <-timeout:
			stream.Stop()
			ch = nil
			break OUTER
		}
	}
}
