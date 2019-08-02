package dcmi

import (
	"context"

	"github.com/gebn/bmc"
)

type sessionlessCommander struct {
	bmc.Sessionless
}

func (s sessionlessCommander) GetDCMICapabilitiesInfoSupportedCapabilities(ctx context.Context) (*GetDCMICapabilitiesInfoSupportedCapabilitiesRsp, error) {
	cmd := NewGetDCMICapabilitiesInfoSupportedCapabilitiesCmd()
	// technically, DCMI uses a different set of codes, but all we're doing here
	// is checking for CompletionCodeNormal
	return &cmd.Rsp, bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd))
}

func (s sessionlessCommander) GetDCMICapabilitiesInfoMandatoryPlatformAttrs(ctx context.Context) (*GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp, error) {
	cmd := NewGetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd()
	return &cmd.Rsp, bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd))
}

func (s sessionlessCommander) GetDCMICapabilitiesInfoOptionalPlatformAttrs(ctx context.Context) (*GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp, error) {
	cmd := NewGetDCMICapabilitiesInfoOptionalPlatformAttrsCmd()
	return &cmd.Rsp, bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd))
}

func (s sessionlessCommander) GetDCMICapabilitiesInfoManageabilityAccessAttrs(ctx context.Context) (*GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp, error) {
	cmd := NewGetDCMICapabilitiesInfoManageabilityAccessAttrsCmd()
	return &cmd.Rsp, bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd))
}

func (s sessionlessCommander) GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrs(ctx context.Context) (*GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp, error) {
	cmd := NewGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd()
	return &cmd.Rsp, bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd))
}

// NewSessionlessCommander wraps a session-less connection in a context that
// provides high-level access to DCMI commands. For convenience, this function
// accepts the Sessionless interface, however DCMI is unlikely to work over IPMI
// v1.5. When sending repeated commands, it is recommended to use the
// SendCommand() method on the connection directly to reduce the number of
// allocations.
func NewSessionlessCommander(s bmc.Sessionless) SessionlessCommands {
	return &sessionlessCommander{
		Sessionless: s,
	}
}
