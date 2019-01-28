###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2019. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
#
# WARNING: DO NOT MODIFY. Changes may be overwritten in future updates.
#
# The following build goals are designed to be generic for any docker image.
# This Makefile is designed to be included in other Makefiles.
# You must ensure that Make variables are defined for IMAGE_REPO and IMAGE_NAME.
#
# If you are using a Bluemix image registry, you must also define BLUEMIX_API_KEY,
# BLUEMIX_ORG, and BLUEMIX_SPACE
###############################################################################

TAG_VERSION ?= `cat VERSION`+$(GIT_COMMIT)

.DEFAULT_GOAL := all

.PHONY: dep-install
glide-install::
	
	mkdir -p $(GOPATH)/bin
	dep version; \
	if [ $$? -ne 0 ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
	fi

.PHONY: pre-req
pre-req::
	go get -v github.com/jteeuwen/go-bindata/...
	dep ensure -v
	go-bindata -pkg i18nBinData -o api/i18n/i18nBinData/i18nTranslations.go -prefix api/i18n/resources api/i18n/resources/*

.PHONY: go-test
go-test:: 
	go test -p 1 -v github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/...

.PHONY: copyright-check
copyright-check:
	./build-tools/copyright-check.sh

.PHONY: tag
tag::
	$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD))
	@echo "TAG_VERSION:$(TAG_VERSION)"

.PHONY: all
all:: clean pre-req copyright-check go-test server client code

.PHONY: clean
clean::
	rm -rf api/testFile
	
#This requires Graphitz and    ''
.PHONY: dependency-graph-text
dependency-graph-text:
	go get github.com/kisielk/godepgraph
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunnerCLI | sed 's/github.ibm.com\/IBMPrivateCloud\/cfp-commands-runner\/api\///' > crcli-dependency-graph.txt
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner | sed 's/github.ibm.com\/IBMPrivateCloud\/cfp-commands-runner\/api\///' > cm-dependency-graph.txt

.PHONY: dependency-graph
dependency-graph: dependency-graph-text
	cat crcli-dependency-graph.txt | dot -Tpng -o crcli-dependency-graph.png
	cat cm-dependency-graph.txt | dot -Tpng -o cm-dependency-graph.png

.PHONY: server
server:
	mkdir -p examples/_build
	go build -o examples/_build/cr-server  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/server

.PHONY: client
client:
	mkdir -p examples/_build
	go build -o examples/_build/cr-cli  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/client

.PHONY: code
code:
	mkdir -p examples/_build
	go build -o examples/_build/cr-code  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/code

.PHONY: migration
code:
	mkdir -p migrationTools/_build
	go build -o migrationTools/_build/localization  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/migrationTools/convertUIMetadataLocalization
