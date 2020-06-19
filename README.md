# opennos-mgmt

## Run server

```
./opennos-mgmt \
  -bind_address :10161 \
  -config gnmi/openconfig.json \
  -key gnmi/certs/server.key \
  -cert gnmi/certs/server.crt \
  -ca gnmi/certs/ca.crt \
  -username foo \
  -password bar \
  -alsologtostderr
```

## Run client
### GET request
```
gnmi_get \
  -target_addr :10161 \
  -key gnmi/certs/client.key \
  -cert gnmi/certs/client.crt \
  -ca gnmi/certs/ca.crt \
  -target_name server.com \
  -username foo \
  -password bar \
  -alsologtostderr \
  -xpath "/interfaces/interface[name=admin]/config/mtu"
```

```
gnmi_get \
  -target_addr :10161 \
  -key gnmi/certs/client.key \
  -cert gnmi/certs/client.crt \
  -ca gnmi/certs/ca.crt \
  -target_name server.com \
  -username foo \
  -password bar \
  -alsologtostderr \
  -xpath "/interfaces/interface[name=eth-1]/config/mtu" \
  -xpath "/interfaces/interface[name=eth-1]/state/admin-status" \
  -xpath "/interfaces/interface[name=eth-1]/ethernet"
```

### SET request
```
gnmi_set \
  -replace /interfaces/interface[name=admin]/config/mtu:1500 \
  -update /interfaces/interface[name=admin]/config/loopback-mode:true \
  -target_addr :10161 \
  -key gnmi/certs/client.key \
  -cert gnmi/certs/client.crt \
  -ca gnmi/certs/ca.crt \
  -target_name server.com \
  -username foo \
  -password bar \
  -alsologtostderr
```

### Capabilities request
```
gnmi_capabilities \
  -target_addr :10161 \
  -key gnmi/certs/client.key \
  -cert gnmi/certs/client.crt \
  -ca gnmi/certs/ca.crt \
  -target_name server.com \
  -alsologtostderr
```

### TODO
Shouldn't be port breakout responsible for creating and destroying Ethernet interface?
If we create new Aggregate interface and it is STATIC, then send request to Ethernet Management Service.
If we create new Aggregate interface and it is DYNAMIC, then send request to Ethernet Management Service. LACP interface have to be created alongside LAG dynamic request!!!
If we create new Aggregate interface and LAG type is not set, then return invalid configuration.
If create new Ethernet interface then at least speed should be defined as dependency!
If create LAG then check speed of all members!