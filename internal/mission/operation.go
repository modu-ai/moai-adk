package mission

import (
	"context"
	"errors"
	"fmt"
)

type OperationReadback interface {
	OperationApplied(context.Context, OperationReceipt) (bool, error)
}

type OperationInvoker interface {
	Invoke(context.Context, OperationReceipt) error
}

// ReconcileMissionOperation checks authoritative state before every possible
// retry. An observation error is ambiguous and therefore blocks; only an
// explicit not-applied readback permits reusing the same stable operation ID.
func ReconcileMissionOperation(ctx context.Context, receipt OperationReceipt, readback OperationReadback, invoker OperationInvoker) (OperationReceipt, error) {
	if receipt.OperationID == "" || readback == nil || invoker == nil {
		return receipt, errors.New("mission operation: incomplete_reconciliation_boundary")
	}
	if receipt.State == ReceiptReconciled {
		return receipt, nil
	}
	applied, err := readback.OperationApplied(ctx, receipt)
	if err != nil {
		return receipt, fmt.Errorf("mission operation: readback_ambiguous: %w", err)
	}
	if applied {
		receipt.State = ReceiptReconciled
		return receipt, nil
	}
	if err := invoker.Invoke(ctx, receipt); err != nil {
		return receipt, fmt.Errorf("mission operation: invoke: %w", err)
	}
	receipt.State = ReceiptInvoked
	return receipt, nil
}
