################################################################################
#
# opennos-mgmt
#
################################################################################
OPENNOS_MGMT_VERSION = 1.0
OPENNOS_MGMT_SITE = $(BR2_EXTERNAL_DESTINY_PATH)/package/opennos-mgmt
OPENNOS_MGMT_SITE_METHOD = local
OPENNOS_MGMT_LICENSE = Apache-2.0
# OPENNOS_MGMT_LICENSE_FILES = LICENSE
OPENNOS_MGMT_DEPENDENCIES = bcm-eth-switch-mgmt

define OPENNOS_MGMT_POST_RSYNC_HOOK
	# GOPATH=$(@D)/_gopath ${GO_BIN} get -u github.com/sirupsen/logrus
	mkdir -p $(@D)/_gopath/{bin,pkg,src}
	cp -Rf $(BCM_ETH_SWITCH_MGMT_DIR)/_gopath/src/* $(@D)/_gopath/src
	GOPATH=$(@D)/_gopath ${GO_BIN} get -u github.com/openconfig/gnmi/cmd/gnmi_cli
	GOPATH=$(@D)/_gopath ${GO_BIN} get -u github.com/jinzhu/copier
	GOPATH=$(@D)/_gopath ${GO_BIN} get -u github.com/r3labs/diff
	mkdir -p $(@D)/_gopath/src/github.com/google/
	(cd $(@D)/_gopath/src/github.com/google/ && git clone https://github.com/google/gnxi.git)
	(cd $(@D)/_gopath/src/github.com/google/gnxi/gnmi_get; \
		GOPATH=$(@D)/_gopath ${GO_BIN} build ./gnmi_get.go && GOPATH=$(@D)/_gopath ${GO_BIN} install; \
		cd ../gnmi_set/; \
		GOPATH=$(@D)/_gopath ${GO_BIN} build ./gnmi_set.go && GOPATH=$(@D)/_gopath ${GO_BIN} install; \
		cd ../gnmi_target/; \
		GOPATH=$(@D)/_gopath ${GO_BIN} build ./gnmi_target.go && GOPATH=$(@D)/_gopath ${GO_BIN} install; \
		cd ../gnmi_capabilities/; \
		GOPATH=$(@D)/_gopath ${GO_BIN} build ./gnmi_capabilities.go && GOPATH=$(@D)/_gopath ${GO_BIN} install \
	)
	GOPATH=$(@D)/_gopath ${GO_BIN} get -u github.com/abiosoft/ishell
	mkdir -p $(@D)/_gopath/src/opennos-mgmt
	mv $(@D)/main.go $(@D)/_gopath/src/opennos-mgmt
	mv $(@D)/management $(@D)/_gopath/src/opennos-mgmt
	mv $(@D)/gnmi $(@D)/_gopath/src/opennos-mgmt
	mv $(@D)/utils $(@D)/_gopath/src/opennos-mgmt
endef

OPENNOS_MGMT_POST_RSYNC_HOOKS += OPENNOS_MGMT_POST_RSYNC_HOOK

define OPENNOS_MGMT_BUILD_CMDS
	cd $(@D)/_gopath/src/opennos-mgmt; \
	CGO_CFLAGS=-I/workdir/buildconfig/br_output/host/x86_64-buildroot-linux-gnu/sysroot/usr/include/bcm-opennsl/ \
	GO111MODULE=off \
	GOARCH=amd64 \
	GOCACHE="/workdir/buildconfig/br_output/host/usr/share/go-cache" \
	GOROOT="/workdir/buildconfig/br_output/host/lib/go" \
	CC="/workdir/buildconfig/br_output/host/bin/x86_64-unknown-linux-gnu-gcc" \
	CXX="/workdir/buildconfig/br_output/host/bin/x86_64-unknown-linux-gnu-g++" \
	GOTOOLDIR="/workdir/buildconfig/br_output/host/lib/go/pkg/tool/linux_amd64" \
	PATH="/workdir/buildconfig/br_output/host/bin:/workdir/buildconfig/br_output/host/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
	GOBIN= CGO_ENABLED=1 \
	GOPATH=$(@D)/_gopath ${GO_BIN} build -o opennos-mgmt
endef

define OPENNOS_MGMT_INSTALL_TARGET_CMDS
	cp $(@D)/_gopath/src/opennos-mgmt/opennos-mgmt $(TARGET_DIR)/usr/bin
	cp $(@D)/_gopath/bin/gnmi_get $(TARGET_DIR)/usr/bin
	cp $(@D)/_gopath/bin/gnmi_set $(TARGET_DIR)/usr/bin
	cp $(@D)/_gopath/bin/gnmi_capabilities $(TARGET_DIR)/usr/bin
	cp $(@D)/_gopath/src/opennos-mgmt/gnmi/certs/* $(TARGET_DIR)/etc/ssl/certs/
endef

$(eval $(generic-package))