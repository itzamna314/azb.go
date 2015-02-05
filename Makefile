SHELL := /bin/bash

version := $(shell cat VERSION)
project := azb-$(version)
tmp_dir := .artifacts/tmp
proj_tmp_dir := $(tmp_dir)/$(project)

.PHONY: all build clean install destroy archive

all: build

install:
	goop install

build:
	goop go build src/main/azb.go
	mkdir -p bin/ && mv azb bin/

clean:
	goop go clean

destroy: clean
	rm Goopfile.lock
	cd .vendor && rm -rf * && cd .. && rmdir .vendor

archive: clean build
	if [ -d $(tmp_dir) ]; then cd $(tmp_dir) && rm -rf *; fi;
	mkdir -p $(proj_tmp_dir)
	cp README.md $(proj_tmp_dir)
	cp -R bin $(proj_tmp_dir)
	cd $(tmp_dir) && tar -cvzf ../$(project).tar.gz $(project)
	
