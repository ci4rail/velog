# velog
Application to log vehicle data to csv files to be used for further analysis.

Currently, supports MVB and CAN using a Ci4Rail [IOU03](https://docs.ci4rail.com/edge-solutions/moducop/io-modules/iou03/) or [MIO03](https://docs.ci4rail.com/edge-solutions/modusio/mio03/) module for data acquisition.


## Functionality

The velog application stores messages from the MVB and CAN bus in parallel into separate csv files. The csv files are stored in a configurable directory.

The csv files names begin with `mvb` and `can` followed by a number that is incremented for each new file. Different prefixes can be configured in the config file via the `FileName` property.

A new file is created when the current file reaches the maximum size, which is 4GB on a vfat filesystem. A new file is also created when the application is started, for example when the system is rebooted.

### MVB data acquisition

From the MVB bus, all process data messages (F-Codes 0,1,2,3,4) are acquired. Other messages are ignored.
The velog application then builds an internal object dictionary from the received messages. The object dictionary stores always the latest value of each MVB address.

After a configurable `DumpInterval`, the object dictionary is dumped to the csv file. The `DumpInterval` is 1 second by default. During each dump, only the objects that have changed since the last dump are written to the csv file. However, when a new file is created, all objects are written to the csv file.

The format of the csv file is as follows (example):

| Dump # | Address (hex) | Last Update - TimeSinceStart (us) | Data (hex) | FCode (dec) | Updates (dec) | 2022-12-27 20:32:31 |
| ------ | ------------- | --------------------------------- | ---------- | ----------- | ------------- | ------------------- |
| 0      | 6af           | 536534091409                      | 58585858   | 1           | 530           |
| 0      | 6b0           | 536534091598                      | 59595959   | 1           | 12            |

Where
* `Dump #` is the number of the dump
* `Address (hex)` is the MVB address
* `Last Update - TimeSinceStart (us)` is the time in microseconds since the start of IO module when the last update of the object was received
* `Data (hex)` is the data of the object. It is composed of the data bytes, each byte is represented by 2 hex characters.
* `FCode (dec)` is the F-Code of the object and
* `Updates (dec)` is the number of messages received on that object since the previous dump.

Also note the timestamp in the first row, which is the absolute time when the file was created.

### CAN data acquisition

For CAN no object dictionary is used. The velog application stores all received CAN messages in the csv file. However, a CAN filter can be configured to only store messages that pass the filter. See the `AcceptanceMask` and `AcceptanceCode` properties in the config file.

The format of the csv file is as follows (example):

| TimeSinceStart (us) | ID (hex) | Data (hex)       | Ext | RTR | 2022-12-27 20:32:32 |
| ------------------- | -------- | ---------------- | --- | --- | ------------------- |
| 1054237352328       | 200      | 0000000000000000 | X   |     |
| 1054237352541       | 201      |                  |     | R   |
| 1054237352654       | 202      | 11               |     |     |

Where
* `TimeSinceStart (us)` is the time in microseconds since the start of IO module when the message was received
* `ID (hex)` is the CAN ID
* `Data (hex)` is the data of the message. It is composed of the data bytes, each byte is represented by 2 hex characters.
* `Ext` is `X` if the message has an extended identifier otherwise it is empty
* `RTR` is `R` if the message is a remote transmission request, otherwise it is empty

Also note the timestamp in the first row, which is the absolute time when the file was created.

### Behavior when Disk is Full

When the disk is full, the velog application will stop writing to the csv files.

## Building

Official releases are built via github actions and are available on [the Releases page of this repo](https://github.com/ci4rail/velog/releases).

Local builds can be done using the following commands:

```bash
cd cmd/logger
GOOS=linux GOARCH=arm64 go build -o ../bin/velog
```

## Configuration

The configuration is done via a config file.

By default, the config file is `$HOME/.velog-config.yaml`, but can be changed with the `--config` flag. When run as a systemd service, the config file is usually at `/etc/velog-config.yaml`.

An example can be found [here](cmd/logger/.velog-config.yaml) with the following content:

```yaml
LoggerOutputDir: /media/sdcard
mvb:
  SnifferDevice: S101-IOU03-USB-EXT-1-mvbSniffer
  FileName: mvb
  DumpInterval: 1000
can:
  SnifferDevice: S101-IOU03-USB-EXT-1-can
  FileName: can
  Bitrate: 500000
  SJW: 3
  SamplePoint: 0.8
  AcceptanceMask: 0x00000000
  AcceptanceCode: 0x00000000
```

The `LoggerOutputDir` property specifies the directory where the csv files are stored. The velog application tries to create the directory if it does not exist.

The `mvb` and `can` sections contain the configuration for the MVB and CAN data acquisition, respectively.

If the `mvb` section is not present, no MVB data is acquired.

If the `can` section is not present, no CAN data is acquired.

The `mvb.SnifferDevice` property specifies the name of the MVB sniffer to use. See the [docs](https://docs.ci4rail.com/edge-solutions/moducop/io-modules/iou03/quick-start-guide/#determine-the-service-address-of-your-iou03) for more info.

The `mvb.FileName` property specifies the prefix of the MVB csv file names.

The `mvb.DumpInterval` property specifies the interval in milliseconds at which the MVB object dictionary is dumped to the csv file.

The `can.SnifferDevice` property specifies the name of the CAN sniffer to use. See the [docs](https://docs.ci4rail.com/edge-solutions/moducop/io-modules/iou03/quick-start-guide/#determine-the-service-address-of-your-iou03) for more info.

The `can.FileName` property specifies the prefix of the CAN csv file names.

The `can.Bitrate` property specifies the bitrate of the CAN bus.

The `can.SJW` property specifies the Synchronization Jump Width of the CAN bus.

The `can.SamplePoint` property specifies the Sample Point of the CAN bus.

The `can.AcceptanceMask` and `can.AcceptanceCode` properties specify the CAN filter. See the [docs]([`can.AcceptanceCode`](https://docs.ci4rail.com/edge-solutions/moducop/io-modules/iou03/detailed-description/#controlling-the-stream-1)) for more info.
