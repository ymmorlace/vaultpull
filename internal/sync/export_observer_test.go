package sync

// ObserverFunc is re-exported so external test packages can use the functional
// adapter without importing the internal package directly via an alias.
type ObserverFunc = observerFuncAlias

// observerFuncAlias mirrors ObserverFunc for the export shim.
type observerFuncAlias = ObserverFunc
