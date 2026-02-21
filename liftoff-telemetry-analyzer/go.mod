module github.com/dladlk/liftoff-telemetry-analyzer

require github.com/dladlk/liftoff-auto-drone v1.2.3

require github.com/dladlk/liftoff-telemetry v0.0.0-00010101000000-000000000000

replace github.com/dladlk/liftoff-telemetry => ..

replace github.com/dladlk/liftoff-auto-drone => ../liftoff-auto-drone

go 1.25.6
