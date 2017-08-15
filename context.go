package main

// FixContext is just the app's directory/lccn context so we don't have global
// variables puked out everywhere but we also don't pass a million args around
// to everything
type FixContext struct {
	SourceDir string
	DestDir   string
	BadLCCN   string
	GoodLCCN  string
}
