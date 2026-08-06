# internal-hooks

Routes hook events to registered handlers.

HookDispatcher is the central router: events arrive tagged with their
kind, the dispatcher consults the handler table, and invokes each
registered handler in registration order.
