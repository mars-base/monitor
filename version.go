package main

// Injected at build time via -ldflags
var (
	AppName    = "Monitor"
	AppVersion = "dev"
	BuildTime  = ""
	GitCommit  = ""
)
