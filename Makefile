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
include Configfile

.PHONY: clean
clean::
	rm -rf _build

.PHONY: set-app-version
set-app-version::
	@echo "RELEASE_VERSION:"$(RELEASE_VERSION)
	@echo "$(RELEASE_TAG)" > VERSION
	@echo "["`date`"] Writing "$(RELEASE_NAME)".yml file  -------------------"
	@echo "$$CFP_YAML"
	@echo "$$CFP_YAML" > platform-$(RELEASE_NAME).yml

.PHONY: create-yaml
create-yaml::
	@mkdir -p _build
	@echo "["`date +%s`"] View the _build cache directory  -------------------"
	ls -al ./_build
	@echo "["`date`"] Move the YAML file into place if they changed  -------------------"
	diff platform-$(RELEASE_NAME).yml ./_build/platform-$(RELEASE_NAME).yml; \
	if [ $$? -ne 0 ]; then \
		rm -f ./_build/*; \
		cp platform-$(RELEASE_NAME).yml ./_build/platform-$(RELEASE_NAME).yml; \
	fi
	@echo "["`date`"] View the _build cache directory  -------------------"
	ls -al ./_build

.PHONY: pre-req
pre-req::
	go get -v github.com/jteeuwen/go-bindata/...
	glide --version; \
	if [ $$? -ne 0 ]; then \
		curl https://glide.sh/get | sh; \
	fi
	glide --debug install --strip-vendor

# Generate the UI JSON
.PHONY: generate_ui_json
generate_ui_json:: pre-req
	uname -a | grep "Darwin"; \
	if [ $$? -eq 0 ]; then \
		make spruce-mac; \
	else \
		make spruce-linux; \
	fi
	./bin/spruce json api/resource/ui-cf-deploy-vmware.yml > api/resource/ui-cf-deploy-vmware.json
	./bin/spruce json api/resource/ui-cf-deploy-openstack.yml > api/resource/ui-cf-deploy-openstack.json
	./build-tools/create-bom-extensions-yml.sh api/resource/extensions ibm-extensions.txt
	./build-tools/create-bom-extensions-yml.sh api/test/resource/extensions ibm-test-extensions.txt
	./build-tools/create-resource-manager.sh  api/resource/extensions ibm-extensions.txt api/test/resource/extensions ibm-test-extensions.txt

.PHONY: go-test
go-test:: 
	go test -v github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/commandsRunner/stateManager && \
	go test -v github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/commandsRunner

.PHONY: go-build
go-build:: go-test
	mkdir -p _build/cm/linux_amd64
	env GOOS=linux GOARCH=amd64 go build -o ./_build/cm/linux_amd64/cmserver github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cm
	env GOOS=linux GOARCH=amd64 go build -o ./_build/cm/linux_amd64/cm github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cmcli
#	cp ./_build/linux_amd64/cm ./_build/cm
	mkdir -p _build/cm/darwin_amd64
	env GOOS=darwin GOARCH=amd64 go build -o ./_build/cm/darwin_amd64/cmserver github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cm
	env GOOS=darwin GOARCH=amd64 go build -o ./_build/cm/darwin_amd64/cm github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cmcli
	mkdir -p _build/cm/windows_amd64
	env GOOS=windows GOARCH=amd64 go build -o ./_build/cm/windows_amd64/cm github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cmcli
	mkdir -p _build/cm/windows_386
	env GOOS=windows GOARCH=386 go build -o ./_build/cm/windows_386/cm github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cmcli

.PHONY: verify_bom
verify_bom::
    @if [ "$(TRAVIS_BRANCH)" != "" ]; then \
	  chmod +x ./build-tools/verify_bom.sh && \
	  ./build-tools/verify_bom.sh api/resource/extensions \
	fi

.PHONY: image
image:: clean set-app-version create-yaml generate_ui_json verify_bom go-build

.PHONY: run
run:: app-version
	docker run --rm -it -v $(IMAGE_NAME)-$(RELEASE_TAG):/repo_local $(IMAGE_REPO)/$(IMAGE_NAME):$(RELEASE_TAG)

.PHONY: all
all:: push run

.PHONY: spruce-mac
spruce-mac:
	cd bin && ln -sf `ls spruce-mac*` spruce && chmod 755 spruce && cd ..

.PHONY: spruce-linux
spruce-linux:
	cd bin && ln -sf `ls spruce-linux*` spruce && chmod 755 spruce && cd ..

.PHONY: copyright-check
copyright-check:
	./build-tools/copyright-check.sh

.PHONY: publish-release
publish-release: push
	@echo "Travis branch:"$(TRAVIS_BRANCH)
	@if ([ "$(TRAVIS_BRANCH)" = "master" ] && [ "$(TRAVIS_PULL_REQUEST)" = "false" ]) || [ "$(PUSH_RELEASE_REQUESTED)" = "true" ]; then \
	  echo "Publishing container $(IMAGE_NAME)-$(RELEASE_TAG)"; \
		docker tag $(IMAGE_REPO)/$(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_REPO)/$(IMAGE_NAME):$(RELEASE_TAG); \
		docker push $(IMAGE_REPO)/$(IMAGE_NAME):$(RELEASE_TAG); \
	else \
	  echo "Skipping publish step, PUSH_RELEASE_REQUESTED: $(PUSH_RELEASE_REQUESTED)"; \
	fi

#This requires Graphitz and    ''
.PHONY: dependency-graph-text
dependency-graph-text:
	go get github.com/kisielk/godepgraph
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cmcli | sed 's/github.ibm.com\/IBMPrivateCloud\//cfp-commands-runner-test\/api\///' > cmcli-dependency-graph.txt
	godepgraph  -p github.com,gonum.org,gopkg.in -s github.ibm.com/IBMPrivateCloud//cfp-commands-runner-test/api/cm | sed 's/github.ibm.com\/IBMPrivateCloud\//cfp-commands-runner-test\/api\///' > cm-dependency-graph.txt

.PHONY: dependency-graph
dependency-graph: dependency-graph-text
	cat cmcli-dependency-graph.txt | dot -Tpng -o cmcli-dependency-graph.png
	cat cm-dependency-graph.txt | dot -Tpng -o cm-dependency-graph.png
