package store

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/logicblocks/event-store/types"
)

type SerialisationGuarantee interface {
	LockName(namespace string, target types.StreamIdentifier) string
}

type logGuarantee struct{}
type categoryGuarantee struct{}
type streamGuarantee struct{}

func (logGuarantee) LockName(ns string, _ types.StreamIdentifier) string {
	return ns
}

func (categoryGuarantee) LockName(ns string, t types.StreamIdentifier) string {
	return fmt.Sprintf("%s.%s", ns, t.Category)
}

func (streamGuarantee) LockName(ns string, t types.StreamIdentifier) string {
	return fmt.Sprintf("%s.%s.%s", ns, t.Category, t.Stream)
}

var (
	GuaranteeLog      SerialisationGuarantee = logGuarantee{}
	GuaranteeCategory SerialisationGuarantee = categoryGuarantee{}
	GuaranteeStream   SerialisationGuarantee = streamGuarantee{}
)

func LockDigest(name string) int64 {
	h := sha256.Sum256([]byte(name))
	n := new(big.Int).SetBytes(h[:])
	mod := new(big.Int).SetInt64(1e16)
	return new(big.Int).Mod(n, mod).Int64()
}
