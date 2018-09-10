all:build
.PHONY : all build  help clean run

BUILD_ROOT ?= ../..
include $(BUILD_ROOT)/common.mk

BUILD_VERSION   ?= 1.0.0
BUILD_TIME      := $(shell date "+%F %T %z")
BUILD_NAME      := devices
COMMIT_SHA1     ?= $(shell  git rev-parse HEAD 2>/dev/null)
SOURCE          := ./main.go
OUT_PUT         ?=./output
CONFIG_FILE     :=localcfg.json

SOURCES := $(wildcard ./*.go) $(wildcard ./restful/*.go) $(wildcard ./managers/*.go) $(wildcard ./rc/*.go)
TARGET :=  $(OUT_PUT)/$(BUILD_NAME)


build:$(TARGET) $(OUT_PUT)/$(CONFIG_FILE)

$(TARGET):$(SOURCES)
	docker run --rm  -v "$(shell pwd)":/usr/src/myapp/src -w /usr/src/myapp/src \
	-e GOOS=linux -e GOARCH=amd64 builder/alpine/golang-devices \
	go build -ldflags                       \
	"                                       \
	-X 'main.BuildVersion=$(BUILD_VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)'       \
	-X 'main.BuildName=$(BUILD_NAME)'       \
	-X 'main.CommitID=$(COMMIT_SHA1)'       \
	"                                       \
	-o ${BUILD_NAME} $(SOURCE)
	@mkdir -p $(OUT_PUT);\
	mv -f ./$(BUILD_NAME) $(OUT_PUT)/;
	$(call log_info,"Finish build $(BUILD_NAME)")

$(OUT_PUT)/$(CONFIG_FILE):${CONFIG_FILE}
	cp -f ./$(CONFIG_FILE) $(OUT_PUT)

clean:
	@echo "clean bin..."
	@rm $(OUT_PUT)/$(BUILD_NAME)* -f
	@rm $(OUT_PUT)/$(CONFIG_FILE) -f

run:$(TARGET)
	$(call log_warn,"Can not run $(BUILD_NAME) at local host!")


help:
	@printf "all/build\t- build bin\n"
	@printf "run\t- build and run bin\n"
	@printf "clean \t- clean the bin\n"
