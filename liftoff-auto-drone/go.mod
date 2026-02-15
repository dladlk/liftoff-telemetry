module github.com/dladlk/liftoff-auto-drone

go 1.25.6

require github.com/dladlk/liftoff-telemetry v0.0.0-00010101000000-000000000000

replace github.com/dladlk/liftoff-telemetry => ..

require atomicgo.dev/keyboard v0.2.9

require (
	github.com/containerd/console v1.0.3 // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
)
