ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
APPLICATION_DIR := $(ROOT_DIR)/application

IAM_RPC_DIR := $(APPLICATION_DIR)/iam-rpc
IAM_API_DIR := $(APPLICATION_DIR)/iam-api
SYS_API_DIR := $(APPLICATION_DIR)/sys-api
SYS_RPC_DIR := $(APPLICATION_DIR)/sys-rpc
RESOURCE_API_DIR := $(APPLICATION_DIR)/resource-api
RESOURCE_RPC_DIR := $(APPLICATION_DIR)/resource-rpc
REALTIME_DIR := $(APPLICATION_DIR)/realtime
GATEWAY_DIR := $(APPLICATION_DIR)/gateway

BIN_DIR := $(ROOT_DIR)/bin
STYLE ?= goZero
GOCTL := go tool goctl
GOCTL_MODEL_TEMPLATE := $(ROOT_DIR)/goctl-template

PG_URL ?= postgres://postgres:post123@127.0.0.1:5433/microservice_kit?sslmode=disable
PG_SCHEMA ?= public
IAM_TABLES ?= s_api_permission,s_auth_client,s_menu,s_org,s_role,s_user,m_*
SYS_TABLES ?= s_config,s_dict_data,s_login_log,s_oper_log
RESOURCE_TABLES ?= biz_attachment

.PHONY: help init-tree gen-iam-api gen-sys-api gen-resource-api gen-iam-rpc gen-sys-rpc gen-resource-rpc gen-realtime gen-iam-model-pg gen-sys-model-pg gen-resource-model-pg gen-models verify-gorm-template verify-no-sqlx verify gen-all build-gateway build-iam-api build-sys-api build-resource-api build-iam-rpc build-sys-rpc build-resource-rpc build-realtime build-all run-gateway run-iam-api run-sys-api run-resource-api run-iam-rpc run-sys-rpc run-resource-rpc run-realtime package clean

help: ## Display help
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

init-tree: ## Initialize sys-api/sys-rpc directory tree
	@mkdir -p $(SYS_API_DIR)/api
	@mkdir -p $(SYS_API_DIR)/etc
	@mkdir -p $(SYS_API_DIR)/internal/{config,handler,logic,middleware,svc,types}
	@mkdir -p $(SYS_RPC_DIR)/pb
	@mkdir -p $(SYS_RPC_DIR)/client/sysservice
	@mkdir -p $(SYS_RPC_DIR)/etc
	@mkdir -p $(SYS_RPC_DIR)/internal/{config,logic,model,server,svc}
	@mkdir -p $(BIN_DIR)
	@echo "Initialized application tree under $(APPLICATION_DIR)"

gen-iam-api: ## Generate iam-api service code from iam.api
	@echo "Generating iam-api..."
	@$(GOCTL) api go --api=$(IAM_API_DIR)/iam.api --dir=$(IAM_API_DIR) --style=$(STYLE)

gen-sys-api: ## Generate sys-api service code from sys.api
	@echo "Generating sys-api..."
	@$(GOCTL) api go \
		--api=$(SYS_API_DIR)/sys.api \
		--dir=$(SYS_API_DIR) \
		--style=$(STYLE)

gen-resource-api: ## Generate resource-api service code from resource.api
	@echo "Generating resource-api..."
	@$(GOCTL) api go --api=$(RESOURCE_API_DIR)/resource.api --dir=$(RESOURCE_API_DIR) --style=$(STYLE)

gen-iam-rpc: ## Generate iam-rpc service code from iam.proto
	@echo "Generating iam-rpc..."
	@rm -rf $(IAM_RPC_DIR)/pb/*
	@$(GOCTL) rpc protoc \
		-I=$(IAM_RPC_DIR) \
		--go_out=$(IAM_RPC_DIR)/pb \
		--go_opt=module=github.com/gcc798/microservice-kit/application/iam-rpc/pb \
		--go-grpc_out=$(IAM_RPC_DIR)/pb \
		--go-grpc_opt=module=github.com/gcc798/microservice-kit/application/iam-rpc/pb \
		--zrpc_out=$(IAM_RPC_DIR) \
		--style=$(STYLE) \
		-m \
		$(IAM_RPC_DIR)/iam.proto

gen-sys-rpc: ## Generate sys-rpc service code from sys.proto
	@echo "Generating sys-rpc..."
	@rm -rf $(SYS_RPC_DIR)/pb/*
	@rm -rf $(SYS_RPC_DIR)/client/sysservice $(SYS_RPC_DIR)/internal/server/sysservice
	@$(GOCTL) rpc protoc \
		-I=$(SYS_RPC_DIR) \
		--go_out=$(SYS_RPC_DIR)/pb \
		--go_opt=module=github.com/gcc798/microservice-kit/application/sys-rpc/pb \
		--go-grpc_out=$(SYS_RPC_DIR)/pb \
		--go-grpc_opt=module=github.com/gcc798/microservice-kit/application/sys-rpc/pb \
		--zrpc_out=$(SYS_RPC_DIR) \
		--style=$(STYLE) \
		-m \
		$(SYS_RPC_DIR)/sys.proto
	@rm -f $(SYS_RPC_DIR)/pb.go $(SYS_RPC_DIR)/etc/pb.yaml

gen-resource-rpc: ## Generate resource-rpc service code from resource.proto
	@echo "Generating resource-rpc..."
	@mkdir -p $(RESOURCE_RPC_DIR)/pb
	@rm -rf $(RESOURCE_RPC_DIR)/pb/*
	@rm -rf $(RESOURCE_RPC_DIR)/client/resourceservice $(RESOURCE_RPC_DIR)/internal/server/resourceservice
	@$(GOCTL) rpc protoc \
		-I=$(RESOURCE_RPC_DIR) \
		--go_out=$(RESOURCE_RPC_DIR)/pb \
		--go_opt=module=github.com/gcc798/microservice-kit/application/resource-rpc/pb \
		--go-grpc_out=$(RESOURCE_RPC_DIR)/pb \
		--go-grpc_opt=module=github.com/gcc798/microservice-kit/application/resource-rpc/pb \
		--zrpc_out=$(RESOURCE_RPC_DIR) \
		--style=$(STYLE) \
		-m \
		$(RESOURCE_RPC_DIR)/resource.proto

gen-realtime: ## Generate realtime RPC code from realtime.proto
	@echo "Generating realtime..."
	@mkdir -p $(REALTIME_DIR)/pb
	@rm -rf $(REALTIME_DIR)/pb/*
	@rm -rf $(REALTIME_DIR)/client/realtimeservice $(REALTIME_DIR)/internal/server/realtimeservice
	@$(GOCTL) rpc protoc \
		-I=$(REALTIME_DIR) \
		--go_out=$(REALTIME_DIR)/pb \
		--go_opt=module=github.com/gcc798/microservice-kit/application/realtime/pb \
		--go-grpc_out=$(REALTIME_DIR)/pb \
		--go-grpc_opt=module=github.com/gcc798/microservice-kit/application/realtime/pb \
		--zrpc_out=$(REALTIME_DIR) \
		--style=$(STYLE) \
		-m \
		$(REALTIME_DIR)/realtime.proto

gen-iam-model-pg: ## Generate IAM GORM models from PostgreSQL datasource
	@if [ -z "$(PG_URL)" ]; then \
		echo "PG_URL is required. Example:"; \
		echo "make gen-models PG_URL='postgres://user:pass@127.0.0.1:5432/db?sslmode=disable'"; \
		exit 1; \
	fi
	@mkdir -p $(IAM_RPC_DIR)/internal/model
	@echo "Generating IAM GORM models from PostgreSQL..."
	@$(GOCTL) model pg datasource \
		--url="$(PG_URL)" \
		--schema="$(PG_SCHEMA)" \
		--table "$(IAM_TABLES)" \
		--dir $(IAM_RPC_DIR)/internal/model \
		--strict \
		--home=$(GOCTL_MODEL_TEMPLATE)

gen-sys-model-pg: ## Generate SYS GORM models from PostgreSQL datasource
	@mkdir -p $(SYS_RPC_DIR)/internal/model
	@echo "Generating SYS GORM models from PostgreSQL..."
	@$(GOCTL) model pg datasource \
		--url="$(PG_URL)" \
		--schema="$(PG_SCHEMA)" \
		--table "$(SYS_TABLES)" \
		--dir $(SYS_RPC_DIR)/internal/model \
		--strict \
		--home=$(GOCTL_MODEL_TEMPLATE)

gen-resource-model-pg: ## Generate Resource GORM models from PostgreSQL datasource
	@mkdir -p $(RESOURCE_RPC_DIR)/internal/model
	@echo "Generating Resource GORM models from PostgreSQL..."
	@$(GOCTL) model pg datasource \
		--url="$(PG_URL)" \
		--schema="$(PG_SCHEMA)" \
		--table "$(RESOURCE_TABLES)" \
		--dir $(RESOURCE_RPC_DIR)/internal/model \
		--strict \
		--home=$(GOCTL_MODEL_TEMPLATE)

gen-models: gen-iam-model-pg gen-sys-model-pg gen-resource-model-pg ## Generate all service-owned GORM models

verify-gorm-template: ## Generate, compile, and run CRUD against PostgreSQL with the GORM model templates
	@GOCTL="$(GOCTL)" ./scripts/verify-gorm-template.sh

verify-no-sqlx: ## Reject source dependencies on the old go-zero SQLX data layer
	@if rg -n 'go-zero/core/stores/(sqlx|sqlc)' --glob '*.go' .; then \
		echo "go-zero SQLX is forbidden; use GORM"; \
		exit 1; \
	fi

verify: verify-no-sqlx ## Run repository checks that do not require external services
	@go test ./...
	@go vet ./...

gen-all: gen-iam-api gen-sys-api gen-resource-api gen-iam-rpc gen-sys-rpc gen-resource-rpc gen-realtime ## Generate API and RPC code

build-gateway: ## Build gateway binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/gateway $(GATEWAY_DIR)/gateway.go

build-iam-api: ## Build iam-api binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/iam-api $(IAM_API_DIR)/iam.go

build-sys-api: ## Build sys-api binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/sys-api $(SYS_API_DIR)/sys.go

build-resource-api: ## Build resource-api binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/resource-api $(RESOURCE_API_DIR)/resource.go

build-iam-rpc: ## Build iam-rpc binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/iam-rpc $(IAM_RPC_DIR)/iam.go

build-sys-rpc: ## Build sys-rpc binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/sys-rpc $(SYS_RPC_DIR)/sys.go

build-resource-rpc: ## Build resource-rpc binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/resource-rpc $(RESOURCE_RPC_DIR)/resource.go

build-realtime: ## Build realtime binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/realtime $(REALTIME_DIR)/realtime.go

build-all: build-gateway build-iam-api build-sys-api build-resource-api build-iam-rpc build-sys-rpc build-resource-rpc build-realtime ## Build all binaries

run-gateway: ## Run gateway service
	@go run $(GATEWAY_DIR)/gateway.go -f $(GATEWAY_DIR)/etc/gateway.yaml

run-iam-api: ## Run iam-api service
	@go run $(IAM_API_DIR)/iam.go -f $(IAM_API_DIR)/etc/iam-api.yaml

run-sys-api: ## Run sys-api service
	@go run $(SYS_API_DIR)/sys.go -f $(SYS_API_DIR)/etc/sys-api.yaml

run-resource-api: ## Run resource-api service
	@go run $(RESOURCE_API_DIR)/resource.go -f $(RESOURCE_API_DIR)/etc/resource-api.yaml

run-iam-rpc: ## Run iam-rpc service
	@go run $(IAM_RPC_DIR)/iam.go -f $(IAM_RPC_DIR)/etc/iam-rpc.yaml

run-sys-rpc: ## Run sys-rpc service
	@go run $(SYS_RPC_DIR)/sys.go -f $(SYS_RPC_DIR)/etc/sys-rpc.yaml

run-resource-rpc: ## Run resource-rpc service
	@go run $(RESOURCE_RPC_DIR)/resource.go -f $(RESOURCE_RPC_DIR)/etc/resource.yaml

run-realtime: ## Run realtime service
	@go run $(REALTIME_DIR)/realtime.go -f $(REALTIME_DIR)/etc/realtime.yaml

package: build-all ## Package binaries
	@tar -czf $(BIN_DIR)/gozero-services.tar.gz -C $(BIN_DIR) gateway iam-api sys-api resource-api iam-rpc sys-rpc resource-rpc realtime
	@echo "Package created: $(BIN_DIR)/gozero-services.tar.gz"

clean: ## Clean build artifacts
	@rm -rf $(BIN_DIR)
	@echo "Cleaned $(BIN_DIR)"
