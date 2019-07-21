package main

// Describe shows various information about a BMC using the ASF Presence Pong,
// Get Channel Authentication Capabilities, Get System GUID and Get Device ID
// commands.

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gebn/bmc"
	"github.com/gebn/bmc/internal/pkg/transport"
	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/alecthomas/kingpin"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	flgBMCAddr = kingpin.Flag("addr", "IP[:port] of the BMC to describe.").
			Required().
			String()
	flgUsername = kingpin.Flag("username", "The username to connect as.").
			Required().
			String()
	flgPassword = kingpin.Flag("password", "The password of the user to connect as.").
			Required().
			String()
)

func main() {
	kingpin.Parse()

	machine, err := bmc.DialV2(*flgBMCAddr) // TODO change to Dial (need to implement v1.5 sessionless communication...)
	if err != nil {
		log.Fatal(err)
	}
	defer machine.Close()

	log.Printf("connected to %v over IPMI v%v", machine.Address(), machine.Version())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := ASFPresencePongCapabilities(ctx, machine); err != nil {
		fmt.Printf("failed to get presence pong capabilities: %v\n", err)
	}
	if err := ChannelAuthenticationCapabilities(ctx, machine); err != nil {
		fmt.Printf("failed to get channel auth capabilities: %v\n", err)
	}
	if err := SystemGUID(ctx, machine); err != nil {
		fmt.Printf("failed to get system GUID: %v\n", err)
	}

	sess, err := machine.NewSession(ctx, *flgUsername, []byte(*flgPassword))
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close(ctx)

	if err := DeviceID(ctx, sess); err != nil {
		fmt.Printf("failed to get device id: %v\n", err)
	}
}

func ASFPresencePongCapabilities(ctx context.Context, t transport.Transport) error {
	pong, err := presencePing(ctx, t)
	if err != nil {
		return err
	}

	fmt.Println("ASF Presence Pong capabilities:")
	fmt.Printf("\tIPMI:               %v\n", pong.IPMI)
	fmt.Printf("\tASF v1.0:           %v\n", pong.ASFv1)
	fmt.Printf("\tASF security exts:  %v\n", pong.SecurityExtensions) // means the BMC uses the secure port in addition to the normal one
	fmt.Printf("\tDASH:               %v\n", pong.DASH)
	fmt.Printf("\tDCMI:               %v\n", pong.SupportsDCMI())
	return nil
}

func presencePing(ctx context.Context, t transport.Transport) (*layers.ASFPresencePong, error) {
	asfRmcp := &layers.RMCP{
		Version:  layers.RMCPVersion1,
		Sequence: 0xFF, // do not send an ACK
		Class:    layers.RMCPClassASF,
	}
	asf := &layers.ASF{
		ASFDataIdentifier: layers.ASFDataIdentifierPresencePing,
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	if err := gopacket.SerializeLayers(buf, opts, asfRmcp, asf); err != nil {
		return nil, err
	}
	bytes, err := t.Send(ctx, buf.Bytes())
	if err != nil {
		return nil, err
	}
	packet := gopacket.NewPacket(bytes, layers.LayerTypeRMCP, gopacket.DecodeOptions{
		Lazy:   true,
		NoCopy: true,
	})
	pongLayer := packet.Layer(layers.LayerTypeASFPresencePong)
	if pongLayer == nil {
		return nil, fmt.Errorf("no presence pong layer in response")
	}
	return pongLayer.(*layers.ASFPresencePong), nil
}

func ChannelAuthenticationCapabilities(ctx context.Context, s bmc.Sessionless) error {
	caps, err := s.GetChannelAuthenticationCapabilities(ctx,
		&ipmi.GetChannelAuthenticationCapabilitiesReq{
			ExtendedData:      true, // only has effect if v2.0
			Channel:           ipmi.ChannelPresentInterface,
			MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
		})
	if err != nil {
		return err
	}
	fmt.Println("Channel Authentication Capabilities:")
	fmt.Printf("\tChannel:            %v\n", caps.Channel)
	fmt.Printf("\tExtended:           %v\n", caps.ExtendedCapabilities)
	fmt.Printf("\tSupportsV2:         %v\n", caps.SupportsV2)
	fmt.Printf("\tK_G configured:     %v\n", caps.TwoKeyLogin)
	fmt.Printf("\tPer-message auth:   %v\n", caps.PerMessageAuthentication)
	fmt.Printf("\tUser-level auth:    %v\n", caps.UserLevelAuthentication)
	fmt.Printf("\tNon-null usernames: %v\n", caps.NonNullUsernamesEnabled)
	fmt.Printf("\tNull usernames:     %v\n", caps.NullUsernamesEnabled)
	fmt.Printf("\tAnon login:         %v\n", caps.AnonymousLoginEnabled)
	fmt.Printf("\tOEM:                %v\n", caps.OEM)
	return nil
}

func SystemGUID(ctx context.Context, s bmc.Sessionless) error {
	guid, err := s.GetSystemGUID(ctx)
	if err != nil {
		return err
	}
	buf := [36]byte{}
	encodeHex(buf[:], guid)
	fmt.Println("System:")
	fmt.Printf("\tGUID:               %v\n", string(buf[:]))
	return nil
}

func encodeHex(dst []byte, guid [16]byte) {
	hex.Encode(dst, guid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], guid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], guid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], guid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], guid[10:])
}

func DeviceID(ctx context.Context, s bmc.Session) error {
	id, err := s.GetDeviceID(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Device:")
	fmt.Printf("\tID:                 %v\n", id.ID)
	fmt.Printf("\tRevision:           %v\n", id.Revision)
	fmt.Printf("\tManufacturer:       %v\n", id.Manufacturer)
	fmt.Printf("\tProduct:            %v\n", id.Product)
	fmt.Printf("\tFirmware (major):   %v\n", id.MajorFirmwareRevision)
	fmt.Printf("\tFirmware (minor):   %v\n", id.MinorFirmwareRevision)
	fmt.Printf("\tFirmware (aux):     %v\n", hex.EncodeToString(id.AuxiliaryFirmwareRevision[:]))
	fmt.Printf("\tFirmware:           %v\n", bmc.FirmwareVersion(id))
	return nil
}
