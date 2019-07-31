package dcmi

import (
	"context"
)

// SessionlessCommands contains the high-level API for commands that can be
// executed outside the context of a session.
type SessionlessCommands interface {
	GetDCMICapabilitiesInfoSupportedCapabilities(context.Context) (*GetDCMICapabilitiesInfoSupportedCapabilitiesRsp, error)
	GetDCMICapabilitiesInfoMandatoryPlatformAttrs(context.Context) (*GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp, error)
	GetDCMICapabilitiesInfoOptionalPlatformAttrs(context.Context) (*GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp, error)
	GetDCMICapabilitiesInfoManageabilityAccessAttrs(context.Context) (*GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp, error)
	GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrs(context.Context) (*GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp, error)
}
