module github.com/dladlk/liftoff-telemetry-analyzer

require github.com/dladlk/liftoff-auto-drone v1.2.3

require (
	atomicgo.dev/keyboard v0.2.9 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/dladlk/liftoff-telemetry v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
)

replace github.com/dladlk/liftoff-telemetry => ..

replace github.com/dladlk/liftoff-auto-drone => ../liftoff-auto-drone

go 1.25.6
