package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// To capture a pcap file on linux:
// sudo tcpdump -w ipmidump.pcap -i any "port 623"
func run() error {
	if len(os.Args) < 2 {
		return errors.New("missing required argument: path to .pcap file to decrypt")
	}
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	password, ok := os.LookupEnv("IPMI_PASSWORD")
	if !ok {
		return errors.New("missing required IPMI_PASSWORD env variable")
	}

	handle, err := pcapgo.NewReader(file)
	if err != nil {
		return err
	}
	ph := packetHandler{
		Password: password,
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		ph.packetCountTotal++

		err := ph.handle(packet)
		if err != nil {
			fmt.Println(ph.packetCountTotal, "handle error:", err)
		}
	}
	return nil
}

type packetHandler struct {
	ExpectedUsername string
	Password         string
	OpenSessionRsp   *ipmi.OpenSessionRsp
	RAKPMessage1     *ipmi.RAKPMessage1
	RAKPMessage2     *ipmi.RAKPMessage2
	cipherLayer      *ipmi.AES128CBC
	decode           gopacket.DecodingLayerFunc

	packetCountTotal int
}

func (p *packetHandler) handle(packet gopacket.Packet) error {

	ignoredLayerTypes := []gopacket.LayerType{
		ipmi.LayerTypeOpenSessionReq, // not decodable yet
		ipmi.LayerTypeRAKPMessage3,   // not decodable yet
		ipmi.LayerTypeRAKPMessage4,
	}
	for _, t := range ignoredLayerTypes {
		layer := packet.Layer(t)
		if layer == nil {
			continue
		}
		fmt.Println(p.packetCountTotal, layer.LayerType())
		return nil
	}

	if layer := packet.Layer(ipmi.LayerTypeOpenSessionRsp); layer != nil {
		p.OpenSessionRsp = layer.(*ipmi.OpenSessionRsp)
		p.RAKPMessage1 = nil
		p.RAKPMessage2 = nil

		fmt.Println(p.packetCountTotal, "Open Session Response")

		switch authAlgo := p.OpenSessionRsp.AuthenticationPayload.Algorithm; authAlgo {
		case ipmi.AuthenticationAlgorithmHMACSHA1:
		default:
			return fmt.Errorf("unsupported authentication algorithm: %s", authAlgo.String())
		}

		switch integAlgo := p.OpenSessionRsp.IntegrityPayload.Algorithm; integAlgo {
		case ipmi.IntegrityAlgorithmHMACSHA196:
		default:
			return fmt.Errorf("unsupported integrity algorithm: %s", integAlgo.String())
		}
		return nil
	}

	if layer := packet.Layer(ipmi.LayerTypeRAKPMessage1); layer != nil {
		fmt.Println(p.packetCountTotal, "RAKP Message 1")
		if p.OpenSessionRsp == nil {
			return errors.New("got RAKP Message 1 before the open session request")
		}
		p.RAKPMessage1 = layer.(*ipmi.RAKPMessage1)
		if p.ExpectedUsername != "" && p.RAKPMessage1.Username != p.ExpectedUsername {
			return fmt.Errorf("unexpected username; expected %q, got %q", p.ExpectedUsername, p.RAKPMessage1.Username)
		}
		return nil
	}

	if layer := packet.Layer(ipmi.LayerTypeRAKPMessage2); layer != nil {
		if p.RAKPMessage1 == nil {
			return errors.New("got RAKP Message 2 before the RAKP Message 1")
		}
		p.RAKPMessage2 = layer.(*ipmi.RAKPMessage2)

		hashGenerator, err := algorithmAuthenticationHashGenerator(p.OpenSessionRsp.AuthenticationPayload.Algorithm)
		if err != nil {
			return err
		}

		effectiveBMCKey := make([]byte, 16)
		copy(effectiveBMCKey, []byte(p.Password))

		sikHash := hashGenerator.SIK(effectiveBMCKey)
		sik := calculateSIK(sikHash, p.RAKPMessage1, p.RAKPMessage2)

		k2Hash := hashGenerator.K(sik)
		k2Hash.Write(bytes.Repeat([]byte{0x02}, 20))
		k2 := k2Hash.Sum(nil)

		key := [16]byte{}
		copy(key[:], k2)
		fmt.Printf("%d RAKP Message 2 Key[% x]\n", p.packetCountTotal, key)

		p.cipherLayer, err = ipmi.NewAES128CBC(key)
		if err != nil {
			return err
		}

		// There is surely a way to include the key in the packet stack
		// so that gopacket can decrypt and decode them.
		// But I have no idea how...

		// keyMaterialGen := additionalKeyMaterialGenerator{
		// 	hash: hashGenerator.K(sik),
		// }

		// cipherLayer, err := algorithmCipher(
		// 	p.OpenSessionRsp.ConfidentialityPayload.Algorithm, keyMaterialGen)
		// if err != nil {
		// 	return err
		// }
		// p.cipherLayer = cipherLayer

		// dlc := gopacket.DecodingLayerContainer(gopacket.DecodingLayerArray(nil))
		// dlc = dlc.Put(cipherLayer)
		// dlc = dlc.Put(&ipmi.Message{})
		// p.decode = dlc.LayersDecoder(ipmi.LayerTypeV2Session, gopacket.NilDecodeFeedback)

		return nil
	}

	if p.cipherLayer == nil {
		return errors.New("no cipherLayer set yet")
	}

	// quick and dirty: get the encrypted payload and decypher it manually
	// print the decoded payload
	if layer := packet.Layer(ipmi.LayerTypeSessionSelector); layer != nil {
		sess := layer.(*ipmi.SessionSelector)

		encrypted := sess.Payload[12:]
		if len(encrypted) < 2*16 {
			return errors.New("payload too short")
		}
		encrypted = encrypted[:2*16]
		err := p.cipherLayer.DecodeFromBytes(encrypted, nil)
		if err != nil {
			return fmt.Errorf("Payload[% x]\n%w", encrypted, err)
		}

		fmt.Printf("%d Decoded[% x]\n", p.packetCountTotal, p.cipherLayer.Payload)

		return nil
	}

	if p.cipherLayer == nil {
		return errors.New("no cipherLayer set yet")
	}

	return nil
}
