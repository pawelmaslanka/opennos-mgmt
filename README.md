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
Withdrawing breakout port mode should depend on if slave ports still exists
Shouldn't be port breakout responsible for creating and destroying Ethernet interface?
During removing Ethernet interface there should be check all dependencies from this interface
Before disable port breakout check if all slave interfaces are removed
Before enable port breakout check if master port has been removed