# Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: go.fmt go.unconvert go.lint go.binaries go.test

GO_PACKAGES := $(shell go list -f '{{.Dir}}' ./...)

go.fmt:
	gofmt -w -s $(GO_PACKAGES)

go.unconvert:
	go run github.com/mdempsky/unconvert -apply -safe $(GO_PACKAGES)

go.lint:
	go run github.com/mgechev/revive -config=.revive.toml -formatter=stylish $(GO_PACKAGES)

go.binaries:
	go run ./cmd/ttn-lw-migrate -h

go.test:
	go test ./... -timeout=5m

go.generate:
	go generate ./...
