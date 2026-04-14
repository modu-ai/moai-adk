// Package aggregator provides a multi-server LSP diagnostic aggregator
// that queries language server clients in parallel, caches results with TTL,
// and degrades gracefully when individual servers fail or time out.
//
// Basic usage:
//
//	mgr := core.NewManager(cfg)
//	agg := aggregator.NewAggregator(mgr)
//	ctx := context.Background()
//	_ = agg.Start(ctx)
//	diags, err := agg.GetDiagnostics(ctx, "/path/to/file.go")
//	_ = agg.Shutdown(ctx)
package aggregator
