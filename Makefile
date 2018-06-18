###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
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
	
	dep ensure -v

.PHONY: go-test
go-test:: 
	go test -v github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner/...

.PHONY: copyright-check
copyright-check:
	./build-tools/copyright-check.sh

.PHONY: all
all:: pre-req copyright-check go-test

#This requires Graphitz and    ''
.PHONY: dependency-graph-text
dependency-graph-text:
	go get github.com/kisielk/godepgraph
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud/commands-runner/api/cmcli | sed 's/github.ibm.com\/IBMPrivateCloud\/commands-runner\/api\///' > cmcli-dependency-graph.txt
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud/cfp-commands-runner/api/commandsRunner | sed 's/github.ibm.com\/IBMPrivateCloud\//cfp-commands-runner\/api\///' > cm-dependency-graph.txt

.PHONY: dependency-graph
dependency-graph: dependency-graph-text
	cat cmcli-dependency-graph.txt | dot -Tpng -o cmcli-dependency-graph.png
	cat cm-dependency-graph.txt | dot -Tpng -o cm-dependency-graph.png

.PHONY: server
server:
	go build -o server  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/server
