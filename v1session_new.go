package bmc

import (
	"context"
	"fmt"
)

func (s *V1SessionlessTransport) NewSession(
	ctx context.Context,
	username string,
	password []byte,
) (Session, error) {
	return nil, fmt.Errorf("not implemented")
}
