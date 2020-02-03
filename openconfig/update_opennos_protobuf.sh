#!/bin/bash

clean() {
  rm -rf public
}

# Ensure that the .pb.go has been generated for the extensions
# that are required.
(cd ${GOPATH}/src/github.com/openconfig/ygot/proto/yext && go generate)
(cd ${GOPATH}/src/github.com/openconfig/ygot/proto/ywrapper && go generate)

clean

OC_PUBLIC_YANG=./public
OC_PUBLIC_RELEASE_MODELS_YANG=${OC_PUBLIC_YANG}/release/models

go run ${GOPATH}/src/github.com/openconfig/ygot/proto_generator/protogenerator.go \
  -generate_fakeroot \
  -base_import_path="." \
  -path=${OC_PUBLIC_YANG}/third_party/ietf,${OC_PUBLIC_RELEASE_MODELS_YANG} -output_dir=ocproto \
  -enum_package_name=enums -package_name=openconfig \
  -exclude_modules=ietf-interfaces \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/platform/openconfig-platform-transceiver.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/lacp/openconfig-lacp.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/lldp/openconfig-lldp.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/stp/openconfig-spanning-tree.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/interfaces/openconfig-interfaces.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/interfaces/openconfig-if-ip.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/interfaces/openconfig-if-aggregate.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/interfaces/openconfig-if-ethernet.yang \
    ${OC_PUBLIC_RELEASE_MODELS_YANG}/interfaces/openconfig-if-ip-ext.yang

go get -u github.com/google/protobuf
proto_imports=".:./ocproto:${GOPATH}/src/github.com/google/protobuf/src:${GOPATH}/src"
find ocproto -name "*.proto" | while read l; do
  protoc -I=$proto_imports --go_out=. $l
done

clean
