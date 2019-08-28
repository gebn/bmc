package bmc

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
)

var (
	defaultAuthenticationAlgorithms = []ipmi.AuthenticationAlgorithm{
		//ipmi.AuthenticationAlgorithmNone,
		ipmi.AuthenticationAlgorithmHMACSHA1,
		//ipmi.AuthenticationAlgorithmHMACMD5,
		//ipmi.AuthenticationAlgorithmHMACSHA256,
	}
	defaultIntegrityAlgorithms = []ipmi.IntegrityAlgorithm{
		//ipmi.IntegrityAlgorithmNone,
		ipmi.IntegrityAlgorithmHMACSHA196,
		//ipmi.IntegrityAlgorithmHMACMD5128,
		//ipmi.IntegrityAlgorithmMD5128,
		//ipmi.IntegrityAlgorithmHMACSHA256128,
	}
	defaultConfidentialityAlgorithms = []ipmi.ConfidentialityAlgorithm{
		//ipmi.ConfidentialityAlgorithmNone,
		ipmi.ConfidentialityAlgorithmAESCBC128,
	}
)

// V2SessionOpts contains configurable parameters for RMCP+ session
// establishment.
type V2SessionOpts struct {
	SessionOpts

	// PrivilegeLevelLookup indicates whether to use both the MaxPrivilegeLevel
	// and Username to search for the relevant user entry. If this is true and
	// the username is empty, we effectively use role-based authentication. If
	// this is false, the supplied MaxPrivilegeLevel will be ignored when
	// searching for the Username.
	PrivilegeLevelLookup bool

	// KG is the key-generating key or "BMC key". It is almost always unset, as
	// it effectively adds a second password in addition to the user/role
	// password, which must be known a-priori to establish a session. It is a 20
	// byte value. If this field is unset, K_[UID], i.e. the user password, will
	// be used in its place (and it is recommended for all 20 bytes of that
	// password to be used to preserve the complexity).
	KG []byte

	// AuthenticationAlgorithms is a slice of authentication algorithms to
	// propose. If this is unspecified, all supported algorithms will be
	// proposed.
	AuthenticationAlgorithms []ipmi.AuthenticationAlgorithm

	// IntegrityAlgorithms is a slice of integrity algorithms to propose for
	// packet signing. If this is unspecified, all supported algorithms will be
	// proposed.
	IntegrityAlgorithms []ipmi.IntegrityAlgorithm

	// ConfidentialityAlgorithms is a slice of confidentiality algorithms to
	// propose for packet encryption. If this is unspecified, all supported
	// algorithms will be proposed.
	ConfidentialityAlgorithms []ipmi.ConfidentialityAlgorithm
}

// NewSession establishes a new RMCP+ session. Two-key login is assumed to be
// disabled (i.e. KG is null), and all algorithms supported by the library will
// be offered. This should cover the majority of use cases, and is recommended
// unless you know a-priori that a BMC key is set.
func (s *V2SessionlessTransport) NewSession(
	ctx context.Context,
	opts *SessionOpts,
) (Session, error) {
	return s.NewV2Session(ctx, &V2SessionOpts{
		SessionOpts: *opts,
	})
}

// NewV2Session establishes a new RMCP+ session with fine-grained parameters.
// This function does not modify the input options. The caller is responsible
// for knowing that v2.0 is supported.
func (s *V2SessionlessTransport) NewV2Session(ctx context.Context, opts *V2SessionOpts) (*V2Session, error) {
	// all the effort is in establish(); this method exists to provide a single
	// point for incrementing the failure count
	sessionOpenAttempts.Inc()
	sess, err := s.newV2Session(ctx, opts)
	if err != nil {
		sessionOpenFailures.Inc()
		return nil, err
	}
	sessionsOpen.Inc()
	return sess, nil
}

func (s *V2SessionlessTransport) newV2Session(ctx context.Context, opts *V2SessionOpts) (*V2Session, error) {
	if opts.AuthenticationAlgorithms == nil {
		opts.AuthenticationAlgorithms = defaultAuthenticationAlgorithms
	}
	if opts.IntegrityAlgorithms == nil {
		opts.IntegrityAlgorithms = defaultIntegrityAlgorithms
	}
	if opts.ConfidentialityAlgorithms == nil {
		opts.ConfidentialityAlgorithms = defaultConfidentialityAlgorithms
	}

	authenticationPayloads := make([]ipmi.AuthenticationPayload,
		len(opts.AuthenticationAlgorithms))
	for i, algo := range opts.AuthenticationAlgorithms {
		authenticationPayloads[i] = ipmi.AuthenticationPayload{
			Algorithm: algo,
		}
	}

	integrityPayloads := make([]ipmi.IntegrityPayload,
		len(opts.IntegrityAlgorithms))
	for i, algo := range opts.IntegrityAlgorithms {
		integrityPayloads[i] = ipmi.IntegrityPayload{
			Algorithm: algo,
		}
	}

	confidentialityPayloads := make([]ipmi.ConfidentialityPayload,
		len(opts.ConfidentialityAlgorithms))
	for i, algo := range opts.ConfidentialityAlgorithms {
		confidentialityPayloads[i] = ipmi.ConfidentialityPayload{
			Algorithm: algo,
		}
	}

	openSessionRsp, err := s.openSession(ctx, &ipmi.OpenSessionReq{
		MaxPrivilegeLevel:       opts.MaxPrivilegeLevel,
		SessionID:               1,
		AuthenticationPayloads:  authenticationPayloads,
		IntegrityPayloads:       integrityPayloads,
		ConfidentialityPayloads: confidentialityPayloads,
	})
	if err != nil {
		return nil, err
	}

	// RAKP Message 1, 2
	remoteConsoleRandom := [16]byte{}
	if _, err := rand.Read(remoteConsoleRandom[:]); err != nil {
		return nil, err
	}
	rakpMessage1 := &ipmi.RAKPMessage1{
		ManagedSystemSessionID: openSessionRsp.ManagedSystemSessionID,
		RemoteConsoleRandom:    remoteConsoleRandom,
		PrivilegeLevelLookup:   opts.PrivilegeLevelLookup,
		MaxPrivilegeLevel:      opts.MaxPrivilegeLevel,
		Username:               opts.Username,
	}
	rakpMessage2, err := s.rakpMessage1(ctx, rakpMessage1)
	if err != nil {
		return nil, err
	}

	hashGenerator, err := algorithmAuthenticationHashGenerator(
		openSessionRsp.AuthenticationPayload.Algorithm)
	if err != nil {
		return nil, err
	}

	authCodeHash := hashGenerator.AuthCode(opts.Password)
	rakpMessage2AuthCode := calculateRAKPMessage2AuthCode(authCodeHash,
		rakpMessage1, rakpMessage2)
	if !hmac.Equal(rakpMessage2.AuthCode, rakpMessage2AuthCode) {
		return nil, fmt.Errorf("RAKP2 HMAC fail: got %v, want %v (this indicates the BMC is using a different password)",
			hex.EncodeToString(rakpMessage2.AuthCode),
			hex.EncodeToString(rakpMessage2AuthCode))
	}

	effectiveBMCKey := opts.KG
	if len(effectiveBMCKey) == 0 {
		effectiveBMCKey = opts.Password
	}
	sikHash := hashGenerator.SIK(effectiveBMCKey)
	sik := calculateSIK(sikHash, rakpMessage1, rakpMessage2)
	icvHash := hashGenerator.ICV(sik)

	rakpMessage4, err := s.rakpMessage3(ctx, &ipmi.RAKPMessage3{
		Status:                 ipmi.StatusCodeOK,
		ManagedSystemSessionID: openSessionRsp.ManagedSystemSessionID,
		AuthCode: calculateRAKPMessage3AuthCode(
			authCodeHash, rakpMessage1, rakpMessage2),
	})
	if err != nil {
		return nil, err
	}
	rakpMessage4ICV := calculateRAKPMessage4ICV(icvHash, rakpMessage1,
		rakpMessage2)
	if !hmac.Equal(rakpMessage4.ICV, rakpMessage4ICV) {
		return nil, fmt.Errorf("RAKP4 ICV fail: got %v, want %v",
			hex.EncodeToString(rakpMessage4.ICV),
			hex.EncodeToString(rakpMessage4ICV))
	}

	keyMaterialGen := additionalKeyMaterialGenerator{
		hash: hashGenerator.K(sik),
	}
	hasher, err := algorithmHasher(openSessionRsp.IntegrityPayload.Algorithm,
		keyMaterialGen)
	if err != nil {
		return nil, err
	}
	cipherLayer, err := algorithmCipher(
		openSessionRsp.ConfidentialityPayload.Algorithm, keyMaterialGen)
	if err != nil {
		return nil, err
	}

	sess := &V2Session{
		v2ConnectionShared:             &s.v2ConnectionShared,
		LocalID:                        openSessionRsp.RemoteConsoleSessionID,
		RemoteID:                       openSessionRsp.ManagedSystemSessionID,
		SIK:                            sik,
		AuthenticationAlgorithm:        openSessionRsp.AuthenticationPayload.Algorithm,
		IntegrityAlgorithm:             openSessionRsp.IntegrityPayload.Algorithm,
		ConfidentialityAlgorithm:       openSessionRsp.ConfidentialityPayload.Algorithm,
		AdditionalKeyMaterialGenerator: keyMaterialGen,
		integrityAlgorithm:             hasher,
		confidentialityLayer:           cipherLayer,
		timeout:                        s.timeout,
	}
	// do not set properties of the session layer here, as it is overwritten
	// each send
	dlc := gopacket.DecodingLayerContainer(gopacket.DecodingLayerArray(nil))
	dlc = dlc.Put(&sess.rmcpLayer)
	dlc = dlc.Put(&sess.sessionSelectorLayer)
	dlc = dlc.Put(&sess.v2SessionLayer)
	dlc = dlc.Put(cipherLayer)
	dlc = dlc.Put(&sess.messageLayer)
	sess.decode = dlc.LayersDecoder(sess.rmcpLayer.LayerType(), gopacket.NilDecodeFeedback)
	return sess, nil
}
