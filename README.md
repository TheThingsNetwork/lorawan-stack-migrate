# lorawan-stack-migrate

Migrate devices from other LoRaWAN Network Servers to [The Things Stack](https://thethingsstack.io).

**IMPORTANT**: `ttn-lw-migrate` is compatible with The Things Stack versions **3.12.0** or newer. Trying to import the devices into earlier versions of The Things Stack will fail, due to breaking API changes.

## Installation

Binaries are available on [GitHub](https://github.com/TheThingsNetwork/lorawan-stack-migrate/releases).

## Support

- [x] The Things Network Stack V2
- [x] [ChirpStack Network Server](https://www.chirpstack.io/)
- [x] [The Things Stack](https://www.github.com/TheThingsNetwork/lorawan-stack/)
- [x] [Firefly](https://fireflyiot.com/)
- [ ] [LORIOT Network Server](https://www.loriot.io/)

Support for different sources is done by creating Source plugins. List available sources with:

```bash
$ ttn-lw-migrate sources
```

## Usage

The `ttn-lw-migrate` examples below export the devices in a `devices.json` file. You will need to import the devices to The Things Stack using this file.

### Console Instructions

Navigate to your application, click **Import End Devices**, select **The Things Stack (JSON)** from the list of available formats, upload the `devices.json` file and click **Create end devices**.

### CLI Instructions

```bash
$ ttn-lw-cli end-devices create --application-id test-app < devices.json
```

## The Things Network Stack V2

### Configuration

Configure with environment variables, or command-line arguments. See `--help` for more details:

```bash
$ export TTNV2_APP_ID="my-ttn-app"                    # TTN App ID
$ export TTNV2_APP_ACCESS_KEY="ttn-account-v2.a..."   # TTN App Access Key (needs `devices` permissions)
$ export FREQUENCY_PLAN_ID="EU_863_870_TTN"           # Frequency Plan for exported devices
```

See [Frequency Plans](https://thethingsstack.io/reference/frequency-plans/) for the list of frequency plans available on The Things Stack. For example, to use `United States 902-928 MHz, FSB 1`, you need to specify the `US_902_928_FSB_1` frequency plan ID.

Private The Things Network Stack V2 deployments are also supported, and require extra configuration. See `ttn-lw-migrate device --help` for more details. For example, to override the discovery server address:

```bash
$ export TTNV2_DISCOVERY_SERVER_ADDRESS="discovery.thethings.network:1900"
```

### Notes

- The export process will halt if any error occurs.
- Execute commands with the `--dry-run` flag to verify whether the outcome will be as expected.
- Payload formatters are not exported. See [Payload Formatters](https://thethingsstack.io/integrations/payload-formatters/).
- For ABP devices, use the `--ttnv2.resets-to-frequency-plan` flag to configure the factory preset frequencies of the device, so that it can keep working with The Things Stack. The list of uplink frequencies is inferred from the Frequency Plan.
- Device sessions (**AppSKey**, **NwkSKey**, **DevAddr**, **FCntUp** and **FCntDown**) are exported by default. You can disable this by using the `--ttnv2.with-session=false` flag. It is recommended that you do not export session keys for devices that can instead re-join on The Things Stack.
- **IMPORTANT**: The migration from The Things Network Stack V2 to The Things Stack is one-way. Note that it is crucial that devices are handled by one Network Server at a time. The commands below will clear both the root keys (**AppKey**, if any) and the session (**AppSKey**, **NwkSKey** and **DevAddr**) from The Things Network Stack V2 after exporting the devices. Make sure you understand the ramifications of this. **Note that having the session keys present on both Network Servers is not supported, and you will most likely encounter uplink/downlink traffic issues and/or a corrupted device MAC state**.

### Export Devices

To export a single device using its Device ID (e.g. `mydevice`):

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate ttnv2 device 'mydevice' --dry-run --verbose > devices.json
# export device
$ ttn-lw-migrate ttnv2 device 'mydevice' > devices.json
```

In order to export a large number of devices, create a file named `device_ids.txt` with one device ID per line:

```
mydevice
otherdevice
device3
device4
device5
```

And then export with:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate ttnv2 devices 'mydevice' --dry-run --verbose < device_ids.txt > devices.json
# export devices
$ ttn-lw-migrate ttnv2 devices < device_ids.txt > devices.json
```

### Export Applications

Similarly, to export all devices of application `my-app-id`:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate ttnv2 application 'my-app-id' --dry-run --verbose > devices.json
# export devices
$ ttn-lw-migrate ttnv2 application 'my-app-id' > devices.json
```

## ChirpStack

### Configuration

Configure with environment variables, or command-line arguments. See `--help` for more details:

```bash
$ export CHIRPSTACK_API_URL="localhost:8080"    # ChirpStack Application Server URL
$ export CHIRPSTACK_API_TOKEN="7F0as987e61..."  # Generate from ChirpStack GUI
$ export JOIN_EUI="0101010102020203"            # JoinEUI for exported devices
$ export FREQUENCY_PLAN_ID="EU_863_870"         # Frequency Plan for exported devices
```

See [Frequency Plans](https://thethingsstack.io/reference/frequency-plans/) for the list of frequency plans available on The Things Stack. For example, to use `United States 902-928 MHz, FSB 1`, you need to specify the `US_902_928_FSB_1` frequency plan ID.

> _NOTE_: `JoinEUI` and `FrequencyPlanID` are required because ChirpStack does not store these fields.

### Notes

- ABP devices without an active session are successfully exported from ChirpStack, but cannot be imported into The Things Stack.
- MaxEIRP may not be always set properly.
- ChirpStack payload formatters also accept a `variables` parameter. This will always be `null` on The Things Stack.

### Export Devices

To export a single device using its DevEUI (e.g. `0102030405060708`):

```
$ ttn-lw-migrate chirpstack device '0102030405060708' > devices.json
```

In order to export a large number of devices, create a file named `device_euis.txt` with one DevEUI per line:

```
0102030405060701
0102030405060702
0102030405060703
0102030405060704
0102030405060705
0102030405060706
```

And then export with:

```bash
$ ttn-lw-migrate chirpstack device < device_euis.txt > devices.json
```

### Export Applications

Similarly, to export all devices of application `chirpstack-app-1`:

```bash
$ ttn-lw-migrate chirpstack application 'chirpstack-app-1' > devices.json
```

In order to export multiple applications, create a file named `application_names.txt` with one Application name per line:

```
chirpstack-app-1
chirpstack-app-2
chirpstack-app-3
```

And export with:

```bash
$ ttn-lw-migrate chirpstack application < application_names.txt > devices.json
```

## The Things Stack

### Configuration

Configure with environment variables, or command-line arguments. See `--help` for more details:

```bash
$ export TTS_APP_ID="my-tts-app"                                                  # TTS App ID
$ export TTS_APP_API_KEY="NNSXS.U..."                                             # TTS App API Key (needs `device` permissions)
$ export TTS_APPLICATION_SERVER_GRPC_ADDRESS="eu1.cloud.thethings.network:8884"   # TTS Application Server URL Address
$ export TTS_IDENTITY_SERVER_GRPC_ADDRESS="eu1.cloud.thethings.network:8884"      # TTS Identity Server URL Address
$ export TTS_JOIN_SERVER_GRPC_ADDRESS="eu1.cloud.thethings.network:8884"          # TTS Join Server URL Address
$ export TTS_NETWORK_SERVER_GRPC_ADDRESS="eu1.cloud.thethings.network:8884"       # TTS Network Server URL Address
$ export TTS_CA_FILE="/path/to/ca.file"                                           # Path to a CA file (optional)
```

### Notes

- The export process will halt if any error occurs.
- Execute commands with the `--dry-run` flag to verify whether the outcome will be as expected.

### Export Device

To export a single device using its Device ID (e.g. `mydevice`):

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate tts device 'mydevice' --dry-run --verbose > devices.json
# export device
$ ttn-lw-migrate tts device 'mydevice' > devices.json
```

In order to export a large number of devices, create a file named `device_ids.txt` with one device ID per line:

```
mydevice
otherdevice
device3
device4
device5
```

And then export with:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate tts devices 'mydevice' --dry-run --verbose < device_ids.txt > devices.json
# export devices
$ ttn-lw-migrate tts devices < device_ids.txt > devices.json
```

### Export Applications

Similarly, to export all devices of application `my-app-id`:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate tts application 'my-app-id' --dry-run --verbose > devices.json
# export devices
$ ttn-lw-migrate tts application 'my-app-id' > devices.json
```

## Firefly

### Configuration

Configure with environment variables, or command-line arguments.

See `ttn-lw-migrate firefly {device|application} --help` for more details.

The following example shows how to set options via environment variables.

```bash
$ export FIREFLY_HOST=example.com       # Host of the Firefly API
$ export FIREFLY_API_KEY=abcdefgh       # Firefly API Key
$ export APP_ID=my-test-app             # Application ID for the exported devices
$ export JOIN_EUI=1111111111111111      # JoinEUI for the exported devices
$ export FREQUENCY_PLAN_ID=EU_863_870   # Frequency Plan ID for the exported devices
$ export MAC_VERSION=1.0.2b             # LoRaWAN MAC version for the exported devices
```

### Notes

- The export process will halt if any error occurs.
- Use the `--invalidate-keys` option to invalidate the root and/or session keys of the devices on the Firefly server. This is necessary to prevent both networks from communicating with the same device. The last byte of the keys will be incremented by 0x01. This enables an easy rollback if necessary. Setting this flag to false (default) would result in a "dry run", where the devices are exported but they will still be able to communicate with the Firefly server.

### Export Devices

To export a single device using its Device EUI (e.g. `1111111111111112`):

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate firefly device 1111111111111112 --verbose > devices.json
# export device
$ ttn-lw-migrate firefly device 1111111111111112 --invalidate-keys > devices.json
```

In order to export a large number of devices, create a file named `device_euis.txt` with one device EUI per line:

```txt
1111111111111112
FF11111111111134
ABCD111111111100
```

And then export with:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate firefly device --verbose < device_ids.txt > devices.json
# export devices
$ ttn-lw-migrate firefly device --invalidate-keys < device_ids.txt > devices.json
```

### Export All Devices

The Firefly LNS does not strictly enforce device to application relationships.

In order to preserve the semantics of the migration tool, the `firefly` source supports the `application` command but in this case, **all devices that are accessible by the API key** are exported.

> Note: Please be cautious while using this command as this might invalidate all the keys of all the devices.

Similarly, to export all devices of application `my-app-id`:

```bash
# dry run first, verify that no errors occur
$ ttn-lw-migrate firefly application all --verbose > devices.json
# export devices
$ ttn-lw-migrate firefly application all --invalidate-keys > devices.json
```

## Development Environment

Requires Go version 1.16 or higher. [Download Go](https://golang.org/dl/).

### Building from source

```bash
$ git clone https://github.com/TheThingsNetwork/lorawan-stack-migrate.git
$ cd lorawan-stack-migrate/
$ go install go.thethings.network/lorawan-stack-migrate/cmd/ttn-lw-migrate
$ $(go env GOPATH)/bin/ttn-lw-migrate --help
```

### Development

Initialize the development environment using `make`:

```bash
$ make init
```

For development/testing purposes, the binary can be executed directly using `go run`:

```bash
$ go run ./cmd/ttn-lw-migrate
```

It is also possible to use `go build`.

## Releasing

### Snapshot releases

Releases are created using [`goreleaser`](https://github.com/goreleaser/goreleaser). First, install GoReleaser:

```bash
$ go install github.com/goreleaser/goreleaser@v1.2.5
```

The command to build a release snapshot from your branch is:

```bash
$ goreleaser --snapshot --rm-dist
```

> Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH`.

This will compile binaries for all supported platforms, `deb`, `rpm`, Snapcraft packages, as well as release archives in `dist`.

> Note: The operating system and architecture represent the name of the directory in `dist` in which the binaries are placed.
> For example, the binaries for Darwin x64 (macOS) will be located at `dist/darwin_amd64`.

### Release from master

1. Create a `release/${version}` branch off the `master` branch.

```bash
$ git checkout master
$ git checkout -b release/${version}
```

2. Update the `CHANGELOG.md` file as explained below:

- Change the **Unreleased** section to the new version and add date obtained via `date +%Y-%m-%d` (e.g. `## [1.0.0] - 2020-10-18`)

  - Check if we didn't forget anything important
  - Remove empty subsections
  - Update the list of links in the bottom of the file
  - Add new **Unreleased** section:

    ```md
    ## [Unreleased]

    ### Added

    ### Changed

    ### Deprecated

    ### Removed

    ### Fixed

    ### Security
    ```

4. Create a pull request targeting `master`.
5. Once this PR is approved and merged, checkout the latest `master` branch locally.
6. Create a version tag, and push to GitHub:

```bash
$ git tag -s -a "v${version}" -m "ttn-lw-migrate v${version}"
$ git push origin "v${version}"
```

7. CI will automatically start building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack-migrate/releases).
8. Edit the release notes on the GitHub releases page, typically copied from `CHANGELOG.md`.
