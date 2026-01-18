package crypto_helpers

import (
	"github.com/VanyaKrotov/xray_cshare/transfer"
	"github.com/xtls/xray-core/common/uuid"
)

const (
	UuidInputOverflow int = 1010
)

func ExecuteUUID(input string) *transfer.Response {
	l := len(input)
	if l == 0 {
		u := uuid.New()

		return transfer.Success(u.String())
	}

	if l <= 30 {
		u, _ := uuid.ParseString(input)

		return transfer.Success(u.String())
	}

	return transfer.New(UuidInputOverflow, "Input must be within 30 bytes.")
}
