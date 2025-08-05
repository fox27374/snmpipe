# SNMPipe
SNMP poller and trap receiver that forwards data to Splunk HEC endpoint.
## Architecture
![architecture](https://github.com/fox27374/snmpipe/blob/20135ce037b06ac1a89084ba5614aefa1bfa7076/doc/architecture.png "Architecture overview")
## Features
* Poll SNMP devices based on a JSON config file
* Every SNMP setting can either be set global or per device
* Scalability through polling go routines
* Data is send as a batch to Splunk HEC
* 
* Structured logging in JSON format
* Implementation is done with the wenn-known gosnmp library
## Limitations
* For SNMPv3, the auth (SHA) and priv (AES) protocols are currently hardcoded
* Only HTTP (not HTTPS) HEC endpoints are supported at the moment
* No MIB support, devices needs to be polled with full OID path
## Configuration
The configuration is done in the **config.json** file. The file has several sections for configuring Splunk, polling settings, trap receiver settings and all the devices that should be polled.
## Usage
### Binary
### Container

