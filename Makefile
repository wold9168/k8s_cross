# Makefile for k8s-cross-cluster project

PLUGIN_NAME=k8s_cross

all: build

# Clone CoreDNS repository to current directory
clone-coredns:
	git submodule update --init --remote;

link-plugin:
	cd coredns/plugin &&\
	if [ ! -e ./"${PLUGIN_NAME}" ]; then \
		ln -s ../../ ./"${PLUGIN_NAME}"; \
	fi


register-plugin: link-plugin
	cd coredns &&\
	grep -qxF ${PLUGIN_NAME}:${PLUGIN_NAME} plugin.cfg || echo ${PLUGIN_NAME}:${PLUGIN_NAME} >> plugin.cfg &&\
	go generate

build: clone-coredns register-plugin
	cd coredns &&\
	make

.PHONY: clone-coredns link-plugin register-plugin build
