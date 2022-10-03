VERSION?="0.0.0"
module := $(shell go list -m)
protoc_include := ${PROTOC_INCLUDE}
protos_path := $(CURDIR)/empirefox/firmata
dart_out := $(CURDIR)/monolith/lib/pb

print:
	echo ${module}

# generate runs `go generate` to build the dynamically generated
# source files, except the protobuf stubs which are built instead with
# "make protobuf".
generate:
	go generate ./...

protobuf-install-plugin:
	@dart pub global activate protoc_plugin
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protobuf:
	@mkdir -p ${dart_out}
	@protoc \
		--proto_path=$(CURDIR) \
		--go_out=paths=import,module=${module}:$(CURDIR) \
		--dart_out=grpc:${dart_out} \
		${protos_path}/*.proto
	@protoc \
		--proto_path=$(CURDIR) \
		--go-grpc_out=paths=import,module=${module}:$(CURDIR) \
		${protos_path}/transport.proto
	@protoc \
		--proto_path=${protoc_include} \
		--dart_out=:${dart_out} \
		${protoc_include}/google/protobuf/empty.proto \
		${protoc_include}/google/protobuf/wrappers.proto

dart-create-project:
	@flutter create monolith --org empirefox.firmata.monolith --project-name monolith

dart-build-runner:
	@cd monolith && flutter pub run build_runner build --delete-conflicting-outputs

dart-pub-get:
	@cd monolith && flutter pub get

jsonschema:
	@protoc \
		--proto_path=$(CURDIR) \
		--jsonschema_out=init/var/planet/schema \
		${protos_path}/board.proto
	@protoc \
		--proto_path=$(CURDIR) \
		--jsonschema_out=init/etc/planet/schema \
		${protos_path}/integration.proto
	@protoc \
		--proto_path=$(CURDIR) \
		--jsonschema_out=init/etc/planet/schema \
		${protos_path}/config.proto

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: print fmtcheck generate protobuf dart-create-project dart-build-runner jsonschema