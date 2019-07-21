package ipmi

import (
//"crypto/aes"
//"crypto/cipher"
//"crypto/rand"
//"fmt"

//"github.com/google/gopacket"
)

//// Cipher is implemented by instances of confidentiality algorithms.
//type Cipher interface {
//
//	// Encrypt converts a plain text IPMI v2.0 session IPMI payload (i.e. an
//	// IPMI message - confidentiality is not defined for other payloads) into
//	// ciphertext, including the correct confidentiality header and/or trailer.
//	// We logically take the payload we would've sent and run it through this
//	// function, replacing it with the result.
//	Encrypt([]byte) ([]byte, error)
//
//	// Decrypt does the opposite of Encrypt. It takes the encrypted message with
//	// confidentiality header and/or trailer, turns it back into plain text, and
//	// returns the raw payload bytes that would have been sent if encryption
//	// were not in use. The result should be deserialisable as an IPMI message.
//	Decrypt([]byte) ([]byte, error)
//}
//
//// noopCipher implements functionality for ConfidentialityAlgorithmNone.
//type noopCipher struct{}
//
//func (n noopCipher) Encrypt(b []byte) ([]byte, error) {
//	return b, nil
//}
//
//func (n noopCipher) Decrypt(b []byte) ([]byte, error) {
//	return b, nil
//}
//
//// aesCBC128Cipher implements the mandatory AES-CBC-128 confidentiality
//// algorithm, whose payload format is specified in 13.29. This is very similar
//// to a manually implemented gopacket layer (but is actually slightly more
//// efficient).
//type aesCBC128Cipher struct {
//	cipher.Block
//}
//
//// K2 is the first 16 bytes of K2, derived from the session SIK as per 13.32
//// of the spec.
//func newAESCBC128Cipher(k2 []byte) (Cipher, error) {
//	block, err := aes.NewCipher(k2[:16])
//	if err != nil {
//		return nil, err
//	}
//	return &aesCBC128Cipher{
//		Block: block,
//	}, nil
//}
//
//func (a *aesCBC128Cipher) Encrypt(b []byte) ([]byte, error) {
//	padLength := 15 - (len(b) % 16)
//	finalLength := 16 + len(b) + padLength + 1
//	payload := make([]byte, finalLength) // single allocation
//
//	// fill first 16 bytes of payload with secure random for IV - becomes
//	// confidentiality header
//	if _, err := rand.Read(payload[:16]); err != nil {
//		return nil, err
//	}
//
//	// write encrypted payload after IV
//	mode := cipher.NewCBCEncrypter(a.Block, payload[:16])
//	mode.CryptBlocks(payload[16:16+len(b)], b)
//
//	// write confidentiality trailer
//	trailerOffset := 16 + len(b)
//	for i := 0; i < padLength; i++ {
//		payload[trailerOffset+i] = uint8(i + 1) // 0x01, 0x02, 0x03 etc.
//	}
//	payload[trailerOffset+padLength] = uint8(padLength)
//	return payload, nil
//}
//
//func (a *aesCBC128Cipher) Decrypt(b []byte) ([]byte, error) {
//	if len(b) < 17 || len(b)%16 != 0 {
//		return nil, fmt.Errorf("AES payload must be at least 17 bytes and have an overall length divisible by 16, got length of %v", len(b))
//	}
//
//	padBytes := uint8(b[len(b)-1])
//	if padBytes > 15 {
//		return nil, fmt.Errorf("invalid number of pad bytes: %v", padBytes)
//	}
//	padStart := len(b) - int(padBytes) - 1
//	// table 13-20 of the spec says we should check the value of each byte of
//	// the pad
//	v := uint8(1)
//	for i := padStart; i < padStart+int(padBytes); i++ {
//		if b[i] != v {
//			return nil, fmt.Errorf("invalid pad byte: offset %v (%v within payload) should have value %v, but has value %v", v-1, i, v, b[i])
//		}
//		v++
//	}
//
//	plaintext := make([]byte, len(b)-16-int(padBytes)-1)
//	mode := cipher.NewCBCDecrypter(a.Block, b[:16])
//	mode.CryptBlocks(plaintext, b[16:16+len(plaintext)])
//
//	return plaintext, nil
//}

//func ipmiPayloadLayerType(c ConfidentialityAlgorithm, s *V2SessionContext) (gopacket.LayerType, error) {
//	buf := [16]byte{}
//	switch c {
//	case ConfidentialityAlgorithmNone:
//		// nothing to do
//		return LayerTypeMessage, nil
//	case ConfidentialityAlgorithmAESCBC128:
//		k2 := s.k(2)
//		copy(buf[:], k2[:16])
//		return NewAESCBC128(buf)
//	//case ConfidentialityAlgorithmXRC4128:
//	//case ConfidentialityAlgorithmXRC440:
//	default:
//		return nil, fmt.Errorf("unsupported confidentiality algorithm: %v", c)
//	}
//}
