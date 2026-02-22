module github.com/dladlk/liftoff-drone-navigator

go 1.25.6

require github.com/dladlk/liftoff-telemetry v0.0.0-00010101000000-000000000000

replace github.com/dladlk/liftoff-telemetry => ..

require github.com/dladlk/liftoff-auto-drone v0.0.0-00010101000000-000000000000

replace github.com/dladlk/liftoff-auto-drone => ../liftoff-auto-drone
