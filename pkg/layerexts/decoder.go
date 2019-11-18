package layerexts

import (
	"github.com/google/gopacket"
)

// layerDecodingLayer is satisfied by types that we can generate a decoder for
// via decodingLayerDecoder. This is lifted from gopacket, where it is not
// exported.
type layerDecodingLayer interface {
	gopacket.Layer
	DecodeFromBytes([]byte, gopacket.DecodeFeedback) error
	NextLayerType() gopacket.LayerType
}

// BuildDecoder creates a gopacket.Decoder for a layer implementing the required
// methods. It is useful when creating a gopacket.LayerTypeMetadata, however
// note this decoder is not used in the context of gopacket.DecodingLayer.
func BuildDecoder(l layerDecodingLayer) gopacket.Decoder {
	return gopacket.DecodeFunc(func(d []byte, p gopacket.PacketBuilder) error {
		err := l.DecodeFromBytes(d, p)
		if err != nil {
			return err
		}
		p.AddLayer(l)
		next := l.NextLayerType()
		if next == gopacket.LayerTypeZero {
			return nil
		}
		return p.NextDecoder(next)
	})
}
