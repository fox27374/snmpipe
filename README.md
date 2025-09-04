# SNMPipe
SNMP poller and trap receiver that forwards data to Splunk HEC endpoint.
## Architecture
![architecture](https://github.com/fox27374/snmpipe/blob/20135ce037b06ac1a89084ba5614aefa1bfa7076/doc/architecture.png "Architecture overview")
## Features
* Poll SNMP devices based on a JSON config file
* SNMP settings can either be set global or per device
* Scalability through go routine polling
* Data is send as batch to Splunk HEC
* Configurable index and sourcetype
* Structured logging in JSON format
* Implementation with the well-known gosnmp library
* Trap receiver and pollin modules can be enabled/disabled
## Limitations
* For SNMPv3, the auth (SHA) and priv (AES) protocols are currently hardcoded
* The certificates for HTTPS HEC enspoints are ignored
* No MIB support, devices needs to be polled with full OID path
## Configuration
The configuration is done in the **config.json** file. The file has several sections for configuring Splunk, polling settings, trap receiver settings and all the devices that should be polled.
```json
"splunk": {
    "splunk_hec_url": "http://hec.example.com:8088/services/collector/event",
    "splunk_hec_token": "xxxxxxxx-d5ce-4950-97a6-yyyyyyyy",
    "splunk_index": "idx_other",
    "splunk_sourcetype": ""
},
"snmp_trap": {
    "enabled": "true",
    "trap_port": "8162",
    "trap_user": "",
    "trap_auth_protocol": "sha",
    "trap_auth_passphrase": "",
    "trap_priv_protocol": "aes",
    "trap_priv_passphrase": ""
},
"snmp_poll": {
    "enabled": "true",
    "interval": "60",
    "snmp_port": "161",
    "snmp_version": "2",
    "snmp_community": "public",
    "snmp_user": "",
    "snmp_auth_protocol": "sha",
    "snmp_auth_passphrase": "",
    "snmp_priv_protocol": "aes",
    "snmp_priv_passphrase": ""
}
```
### splunk
Settings for the Splunk HEC communication. Mandatory settings are **plunk_hec_url** and **plunk_hec_token**. If these are not set, the application is gong to exit.
### snmp_trap
General settings for the trap receiver. There are no mandatory settings. If nothing is set, the defaupt listening port will be **8162**. The application is designed to run as non-root user so be aware that a port number below **1024** is going to need root permissions.
### snmp_poll
General settings for the device polling. No mandatory settings. The **interval** defaults to 60 seconds if nothing is set. Same for the **snmp_port**, which is set to 161.
### devices
```json
"devices": [
    {
        "ip": "172.24.81.200",
        "name": "EATON USV",
        "oids": {
            ".1.3.6.1.4.1.534.1.3.4.1.2.1": "voltage",
            ".1.3.6.1.4.1.534.1.3.4.1.7.1": "current",
            ".1.3.6.1.4.1.534.1.4.9.3.0": "power"
        }
    },
    {
        "ip": "172.24.81.41",
        "name": "Rack 1-1",
        "snmp_community": "internal",
        "snmp_version": "1",
        "oids": {
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.4.1": "voltage",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.5.1": "current",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.6.1": "power",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.10.1": "energy"
        }
    },
    {
        "ip": "172.24.81.42",
        "name": "Rack 1-2",
        "snmp_version": "3",
        "snmp_user": "",
        "snmp_auth_protocol": "sha",
        "snmp_auth_passphrase": "password1234",
        "snmp_priv_protocol": "aes",
        "snmp_priv_passphrase": "password1234"
        "oids": {
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.4.1": "voltage",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.5.1": "current",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.6.1": "power",
            ".1.3.6.1.4.1.318.1.1.30.4.2.1.10.1": "energy"
        }
    }
```
Every device gets the settings from the **snmp_poll** configuration. These settings can be overwritten on a per device basis. The polled values are converted to strings in oder to send them as a JSON payload.
## Usage
### Binary
* Clone the repository
* Rename the **config.json.template** to **config.json**
* Adapt the settings in the config file to your needs
* Then either run the code directly
```bash
go run *.go
```
* Or build the binary and run it
```bash
go build -ldflags="-w -s" .
./snmpipe
```
### Container
Install [Podman](https://podman.io/docs/installation)
```bash
podman machine init
podman machine start
podman buildx build --platform linux/amd64 -t snmpipe:0.2.0 -f container/Containerfile
podman image rm $(podman images -f "dangling=true" -q)
podman run -d --rm --name snmpipe -v $(pwd)/config.json:/etc/snmpipe/config.json -p 8162:8162/udp localhost/snmpipe:0.2.0
```
### Compose
```yaml
services:
  snmpipe:
    image: quay.io/repository/dkofler/snmpipe:0.2.0
    # Alternatively build the image
    # build: container/Containerfile
  environment:
    DEBUG: false
  ports:
    - "8162:8162"
  volumes:
    - ./config.json:/etc/snmpipe/config.json
```
You can also use pre-build images from this [registry](https://quay.io/repository/dkofler/snmpipe?tab=tags). The "latest" tag is always available.
### Kubernetes
TODO
## Troubleshooting
In order to turn on debug logging, the **DEBUG** environmental variable set to **true** can be passed to the application.
### Binary
```bash
DEBUG=true go run *.go
```
### Container
```bash
podman run -d --rm --name snmpipe -e DEBUG=true -v $(pwd)/config.json:/etc/snmpipe/config.json -p 8162:8162/udp localhost/snmpipe:0.2.0
```

