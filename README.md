# cga-led

Small utility to control the LED state of CGA6444VF modems (Vodafone
Station Wi-Fi 6 Technicolor).

## License

This project is licensed under the *GNU General Public License v3.0*. A copy of
the license can be found in [LICENSE](LICENSE).

## Building

If you would like to build the utility for the current machine run the
following command. This will create `bin/cga-led`.

```shell
make build
```

If you would like to build the utility for arm64 machines run the
following command. This will create `bin/cga-led-arm`.

```shell
make build-arm
```

## Usage

Following flags are accepted:

| Flag | Description                              |
|------|------------------------------------------|
| -a   | Address of API (default "192.168.100.1") |
| -u   | Username for API (default "admin")       |
| -p   | Password for API (default "password")    |
| -l   | Turn led on (true) or off (false)        |
