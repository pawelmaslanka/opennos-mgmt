#!/bin/bash
mkdir $GOPATH/src/github.com/golang
cd $GOPATH/src/github.com/golang
git clone https://github.com/golang/protobuf.git
cd protobuf
git checkout git checkout v1.4.0-rc.4 # v1.3.5
go get ./...
make install
##################
mkdir $GOPATH/src/github.com/openconfig
cd $GOPATH/src/github.com/openconfig
git clone https://github.com/openconfig/goyang.git
cd goyang && git checkout 3ecb82e && go get ./... && go build && go install
#####################
mkdir $GOPATH/src/github.com/openconfig
cd $GOPATH/src/github.com/openconfig
git clone https://github.com/openconfig/gnmi.git
cd gnmi && git checkout 58b1f73c2cbd && go get ./...
##############
mkdir $GOPATH/src/google.golang.org
cd $GOPATH/src/google.golang.org
git clone https://github.com/protocolbuffers/protobuf-go.git
mv protobuf-go protobuf
cd protobuf && git checkout v1.20.1
######################
mkdir $GOPATH/src/google.golang.org
cd $GOPATH/src/google.golang.org
git clone https://github.com/googleapis/go-genproto.git
mv go-genproto genproto
cd genproto && git checkout 0848e9f44c36 && go get ./...
######################
mkdir $GOPATH/src/github.com/abiosoft
cd $GOPATH/src/github.com/abiosoft
git clone https://github.com/abiosoft/ishell.git
cd ishell && go get ./...
######################
mkdir $GOPATH/src/github.com/jinzhu
cd $GOPATH/src/github.com/jinzhu
git clone https://github.com/jinzhu/copier.git
cd copier && go get ./...
######################
mkdir $GOPATH/src/github.com/cheekybits
cd $GOPATH/src/github.com/cheekybits
git clone https://github.com/cheekybits/genny.git
cd genny && go get ./...
######################
mkdir $GOPATH/src/github.com/golang
cd $GOPATH/src/github.com/golang
git clone https://github.com/golang/glog.git
cd glog && go get ./...
######################
mkdir $GOPATH/src/github.com/google
cd $GOPATH/src/github.com/google
git clone https://github.com/google/go-cmp.git
cd go-cmp && go get ./...
######################
mkdir $GOPATH/src/github.com/kylelemons
cd $GOPATH/src/github.com/kylelemons
git clone https://github.com/kylelemons/godebug.git
cd godebug && go get ./...
######################
mkdir $GOPATH/src/github.com/r3labs
cd $GOPATH/src/github.com/r3labs
git clone https://github.com/r3labs/diff.git
cd diff && go get ./...
######################
mkdir -p $GOPATH/src/golang.org/x
cd $GOPATH/src/golang.org/x
git clone https://github.com/golang/net.git
cd net && go get ./...
######################
mkdir -p $GOPATH/src/golang.org/x
cd $GOPATH/src/golang.org/x
git clone https://github.com/golang/text.git
cd text && go get ./...
##################
cd $GOPATH/src/github.com/openconfig/ygot/proto/yext && go generate
cd $GOPATH/src/github.com/openconfig/ygot/proto/ywrapper && go generate


# If there is problem with the followig errors:
# github.com/openconfig/ygot/ygot
#../../ygot/proto.go:30:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.BoolValue (missing ProtoReflect method)
#../../ygot/proto.go:32:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.BytesValue (missing ProtoReflect method)
#../../ygot/proto.go:34:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.Decimal64Value (missing ProtoReflect method)
#../../ygot/proto.go:37:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.IntValue (missing ProtoReflect method)
#../../ygot/proto.go:39:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.StringValue (missing ProtoReflect method)
#../../ygot/proto.go:41:5: impossible type switch case: v.Message().Interface() (type protoreflect.ProtoMessage) cannot have dynamic type *ywrapper.UintValue (missing ProtoReflect method)
#interfaces.go:29: running "go": exit status 2
#make: *** [generate] Error 1

# Just run the following commands:
# cd $GOPATH/src/github.com/openconfig/ygot/proto/yext && go generate
# cd $GOPATH/src/github.com/openconfig/ygot/proto/ywrapper && go generate