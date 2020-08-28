# lorawan-stack-migrate

Migrate devices from other LoRaWAN Network Servers to [The Things Stack](https://thethingsstack.io).

## Installation

Binaries are available on [GitHub](https://github.com/TheThingsNetwork/lorawan-stack-migrate/releases).

Docker images are available on [Docker Hub](https://hub.docker.com/r/TheThingsNetwork/lorawan-stack-migrate).

## Building from source

Requires Go version 1.14 or higher. [Download Go](https://golang.org/dl/).

```bash
$ go install go.thethings.network/lorawan-stack-migrate/cmd/ttn-lw-migrate
```

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
