package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type ReceiptState string

const (
	ReceiptPrepared        ReceiptState = "prepared"
	ReceiptInvoked         ReceiptState = "invoked"
	ReceiptObservedApplied ReceiptState = "observed_applied"
	ReceiptReconciled      ReceiptState = "reconciled"
	ReceiptDenied          ReceiptState = "denied"
)

type OperationReceipt struct {
	OperationID  string
	MissionID    string
	DecisionID   string
	Action       Action
	Targets      []string
	SnapshotHash string
	State        ReceiptState
	ReasonCode   string
}

func stableOperationID(contractHash string, decision Decision) string {
	targets := append([]string(nil), decision.Targets...)
	sortStrings(targets)
	marker := strings.Join([]string{decision.MissionID, contractHash, string(decision.Action), strings.Join(targets, "\x00")}, "\x00")
	sum := sha256.Sum256([]byte(marker))
	return "op-" + hex.EncodeToString(sum[:16])
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
