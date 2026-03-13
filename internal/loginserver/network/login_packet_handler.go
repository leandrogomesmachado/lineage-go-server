package network

import (
	"context"
	"fmt"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"
)

func (lc *LoginClient) processarPacket(ctx context.Context, dados []byte) error {
	if len(dados) == 0 {
		return nil
	}

	opcode := dados[0]
	if lc.state == protocol.StateConnected {
		if opcode == protocol.AuthGameGuard {
			return lc.handleAuthGameGuard(dados[1:])
		}
		return fmt.Errorf("opcode 0x%02X nao esperado para estado %s", opcode, lc.state)
	}

	if lc.state == protocol.StateAuthedGG {
		if opcode == protocol.RequestAuthLogin {
			return lc.handleAuthLogin(ctx, dados[1:])
		}
		return fmt.Errorf("opcode 0x%02X nao esperado para estado %s", opcode, lc.state)
	}

	if lc.state == protocol.StateAuthedLogin {
		if opcode == protocol.RequestServerList {
			return lc.handleServerList(ctx, dados[1:])
		}
		if opcode == protocol.RequestServerLogin {
			return lc.handleServerLogin(ctx, dados[1:])
		}
		return fmt.Errorf("opcode 0x%02X nao esperado para estado %s", opcode, lc.state)
	}

	return fmt.Errorf("estado desconhecido: %s", lc.state)
}
