# lorawan-stack-migrate

Migrate devices from other LoRaWAN Network Servers to [The Things Stack](https://thethingsstack.io).

## Installation

Binaries are available on [GitHub](https://github.com/TheThingsNetwork/lorawan-stack-migrate/releases).

Docker images are available on [Docker Hub](https://hub.docker.com/r/TheThingsNetwork/lorawan-stack-migrate).

## Support

- [X] [ChirpStack Network Server](https://www.chirpstack.io/)
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

## ChirpStack

### Configuration

Configure with environment variables, or command-line arguments. See `--help` for more details:

```bash
$ export CHIRPSTACK_API_URL="localhost:8080"    # ChirpStack Application Server URL
$ export CHIRPSTACK_API_TOKEN="7F0as987e61..."  # Generate from ChirpStack GUI
$ export JOIN_EUI="0101010102020203"            # JoinEUI for exported devices
$ export FREQUENCY_PLAN_ID="EU_863_870"         # Frequency Plan for exported devices
```

> *NOTE*: `JoinEUI` and `FrequencyPlanID` are required because ChirpStack does not store these fields.

### Export Devices

To export a single device using its DevEUI (e.g. `0102030405060708`):

```
$ ttn-lw-migrate device --source chirpstack "0102030405060708" > devices.json
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
$ ttn-lw-migrate device --source chirpstack < device_euis.txt > devices.json
```

### Export Applications

Similarly, to export all devices of application `chirpstack-app-1`:

```bash
$ ttn-lw-migrate application --source chirpstack "chirpstack-app-1" > devices.json
```

In order to export multiple applications, create a file named `application_names.txt` with one Application name per line:

```
chirpstack-app-1
chirpstack-app-2
chirpstack-app-3
```

And export with:

```bash
$ ttn-lw-migrate application --source chirpstack < application_names.txt > devices.json
```

### Notes

- ABP devices without an active session are successfully exported from ChirpStack, but cannot be imported into The Things Stack.
- MaxEIRP may not be always set properly.
- ChirpStack payload formatters also accept a `variables` parameter. This will always be `null` on The Things Stack.

## Development Environment

Requires Go version 1.15 or higher. [Download Go](https://golang.org/dl/).

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

Releases are created using [`goreleaser`](https://github.com/goreleaser/goreleaser). You can build a release snapshot from your local branch with `go run github.com/goreleaser/goreleaser --snapshot`.

> Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH`.

This will compile binaries for all supported platforms, `deb`, `rpm` and Snapcraft packages, release archives in `dist`, as well as Docker images.

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
