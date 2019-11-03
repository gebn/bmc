package ipmi

// IntegrityAlgorithm is the identifier of integrity algorithms negotiated in
// the RMCP+ session establishment process. The numbers are defined in 13.28.4
// of the spec. The integrity algorithm is used to calculate the signature for
// authenticated RMCP+ messages.
type IntegrityAlgorithm uint8

const (
	IntegrityAlgorithmNone          IntegrityAlgorithm = iota
	IntegrityAlgorithmHMACSHA196                       // 12 byte authcode
	IntegrityAlgorithmHMACMD5128                       // 16 bytes ''
	IntegrityAlgorithmMD5128                           // 16 bytes ''
	IntegrityAlgorithmHMACSHA256128                    // 16 bytes ''
)

func (i IntegrityAlgorithm) String() string {
	switch i {
	case IntegrityAlgorithmNone:
		return "None"
	case IntegrityAlgorithmHMACSHA196:
		return "HMAC-SHA1-96"
	case IntegrityAlgorithmHMACMD5128:
		return "HMAC-MD5-128"
	case IntegrityAlgorithmMD5128:
		return "MD5-128"
	case IntegrityAlgorithmHMACSHA256128:
		return "HMAC-SHA256-128"
	}
	if 0xc0 <= i && i <= 0xff {
		return "OEM"
	}
	return "Unknown"
}
