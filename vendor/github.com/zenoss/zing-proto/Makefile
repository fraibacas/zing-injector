
ROOTDIR              ?= $(CURDIR)
#
# When running a deploy build in Jenkin ROOTDIR is set to HOST_WORKSPACE, but
# the code is checked out to a "service" subdirectory, so in that case we need
# to adjust the value of ROOTDIR.
#
BASE_CURDIR          := $(shell basename $(CURDIR))
BASE_ROOTDIR         := $(shell basename $(ROOTDIR))
ifneq ($(BASE_CURDIR),$(BASE_ROOTDIR))
    ROOTDIR          := $(ROOTDIR)/$(BASE_CURDIR)
endif

#
# If we're executing in a Jenkins context, use a workspace-local maven repo,
# otherwise use the developer's home directory
#
ifneq  (,$(findstring /var/lib/docker,$(ROOTDIR)))
LOCAL_MAVEN_REPO     = $(ROOTDIR)/.m2
else
LOCAL_MAVEN_REPO     = $(HOME)/.m2
endif

PROTODIR             := protobufs
GODIR                := go
DESCRIPTORDIR        := descriptors
JAVADIR              := java
JAVA_SRC_DIR         := $(JAVADIR)/src/main/java
PROTOFILES           := $(shell find protobufs -name "*.proto")
GOFILES              := $(subst $(PROTODIR), $(GODIR),$(subst .proto,.pb.go,$(PROTOFILES)))
DESCFILES            := $(subst $(PROTODIR), $(DESCRIPTORDIR),$(subst .proto,.descriptor.pb,$(PROTOFILES)))
JAVAFILES            := $(subst $(PROTODIR), $(JAVA_SRC_DIR),$(subst .proto,.java,$(PROTOFILES)))
PROTOC               := /usr/bin/protoc

LOCAL_USER_ID        := $(shell id -u)
CONTAINER_DIR        := /tmp/working
ZENKIT_BUILD_VERSION := 1.6.4
BUILD_IMG            := zenoss/zenkit-build:$(ZENKIT_BUILD_VERSION)
DOCKER_NETWORK       = host
DOCKER_PARAMS        := --rm -v $(ROOTDIR):$(CONTAINER_DIR):rw \
                        -e LOCAL_USER_ID=$(LOCAL_USER_ID) \
                        --network $(DOCKER_NETWORK) \
                        -w $(CONTAINER_DIR)
DOCKER_CMD           := docker run -t $(DOCKER_PARAMS) $(BUILD_IMG)

DEPLOY_BUILD_IMAGE   = zenoss/zing-java-build:1.2
MVN_SETTINGS         ?= /etc/settings.xml
MVN_OPTS             ?= -s $(MVN_SETTINGS)
MVN                  = docker run --rm \
                            --network $(DOCKER_NETWORK) \
                            -v $(ROOTDIR)/$(JAVADIR):/usr/src/app:rw \
                            -v $(LOCAL_MAVEN_REPO):/home/user/.m2:rw \
                            -e LOCAL_USER_ID=$(LOCAL_USER_ID) \
                            -w /usr/src/app \
                            $(DEPLOY_BUILD_IMAGE) \
                            mvn $(MVN_OPTS)

default: all_containerized

.PHONY: all
all: $(GOFILES) $(DESCFILES) $(JAVAFILES)

.PHONY: all_containerized
all_containerized:
	$(DOCKER_CMD) make all
	make package

$(GODIR):
	@mkdir $(GODIR)

$(DESCRIPTORDIR):
	@mkdir $(DESCRIPTORDIR)

$(JAVA_SRC_DIR):
	@mkdir -p $(JAVA_SRC_DIR)

$(GODIR)/%.pb.go: $(PROTOFILES) $(GODIR)
	@$(PROTOC) -I $(PROTODIR) $(PROTODIR)/$*.proto --go_out=plugins=grpc:$(GODIR)

$(DESCRIPTORDIR)/%.descriptor.pb: $(PROTOFILES) $(DESCRIPTORDIR)
	@$(PROTOC) -I $(PROTODIR) $(PROTODIR)/$*.proto --include_source_info --include_imports --descriptor_set_out=$(DESCRIPTORDIR)/$*.descriptor.pb

# This is a PHONY because the actual path of the java files
# will depend on the value of java_package in the .proto file
.PHONY: $(JAVA_SRC_DIR)/%.java
$(JAVA_SRC_DIR)/%.java: $(PROTOFILES) $(JAVA_SRC_DIR)
	@$(PROTOC) -I $(PROTODIR) $(PROTODIR)/$*.proto --java_out=$(JAVA_SRC_DIR)

.PHONY: clean
clean:
	@rm -rf $(GODIR) $(DESCDIR) $(JAVA_SRC_DIR)
	-$(MVN) -e clean

.PHONY: package
package:
	$(MVN) -e clean package

.PHONY: set-relversion
set-relversion:
	$(MVN) versions:set -DnewVersion=$(NEW_VERSION) -DgenerateBackupPoms=false

.PHONY: set-devversion
set-devversion:
	$(MVN) versions:set -DnewVersion=$(NEW_VERSION)-SNAPSHOT -DgenerateBackupPoms=false


# deploy will build the Jar and deploy it to the remote repo without running tests
# The caller is responsible for setting TARGET_MVN_REPO to the appropriate target repo (snapshot or release).
.PHONY: deploy
deploy:
	$(MVN) -e clean deploy \
	-DaltDeploymentRepository=$(TARGET_MVN_REPO) $(MVN_DEPLOY_OPTS)
