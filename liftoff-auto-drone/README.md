# liftoff-auto-drone

Control Liftoff Game drones via simple program

Based on [go-vigem](https://github.com/openstadia/go-vigem) - Go bindings for the ViGEmClient library.

## Requirements

This library requires the [ViGEmClient](https://github.com/ViGEm/ViGEmClient) and [ViGEmBus](https://github.com/ViGEm/ViGEmBus) to work.

- Download and install [ViGEmBus](https://github.com/ViGEm/ViGEmBus/releases)
- Download ViGEmClient.dll
    - [x64](https://buildbot.nefarius.at/builds/ViGEmClient/master/1.21.222.0/bin/release/x64/)

Place ViGEmClient.dll in the same folder as your application. The library should automatically detect dll on startup.

## References

- [go-vigem](https://github.com/openstadia/go-vigem)
- [ViGEm/ViGEmClient](https://github.com/ViGEm/ViGEmClient)
- [ViGEm/ViGEm.NET](https://github.com/ViGEm/ViGEm.NET)
- [yannbouteiller/vgamepad](https://github.com/yannbouteiller/vgamepad)
