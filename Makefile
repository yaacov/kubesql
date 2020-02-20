#
# Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

VERSION_GIT := $(shell git describe --tags)
VERSION ?= ${VERSION_GIT}

kubesql_src := $(wildcard ./cmd/kubesql/*.go)

all: kubesql

kubesql: $(kubesql_src)
	go build -o kubesql -ldflags "-X main.version=${VERSION}" $(kubesql_src)

.PHONY: lint
lint:
	golint $(kubesql_src)

.PHONY: fmt
fmt:
	gofmt -s -w $(kubesql_src)

.PHONY: clean
clean:
	rm -f kubesql
