#!/usr/bin/env bash
# shellcheck disable=SC2086

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TF_DIR="${REPO_ROOT}/project/opentofu"
GENERATED_VARS="${TF_DIR}/generated.auto.tfvars.json"

SERVICES=(
  "api-gateway"
  "authentication-service"
  "location-service"
  "logger-service"
  "trip-service"
)

declare -A SERVICE_CONTEXTS=(
  ["api-gateway"]="api-gateway"
  ["authentication-service"]="authentication-service"
  ["location-service"]="location-service"
  ["logger-service"]="logger-service"
  ["trip-service"]="trip-service"
)

declare -A SERVICE_DOCKERFILES=(
  ["api-gateway"]="api-gateway.dockerfile"
  ["authentication-service"]="authentication-service.dockerfile"
  ["location-service"]="location-service.dockerfile"
  ["logger-service"]="logger-service.dockerfile"
  ["trip-service"]="trip-service.dockerfile"
)

usage() {
  cat <<'EOF'
Usage: deploy.sh [options]

Required options:
  --subscription <id>      Azure subscription ID or name to target
  --resource-group <name>  Resource group to create/update
  --location <region>      Azure region (e.g. eastus)
  --acr-name <name>        Existing or planned Azure Container Registry name
  --key-vault <name>       Existing or planned Azure Key Vault name (global name scope)

Optional arguments:
  --env <name>             Deployment environment (default: dev)
  --tag <tag>              Container image tag (default: current git short SHA)
  --tfvars <file>          Additional tfvars file to merge (default: none)
  --skip-build             Skip container image build/push
  --skip-apply             Run plan only (no apply)
  -h, --help               Show this help message

Examples:
  ./deploy.sh \
    --subscription 00000000-0000-0000-0000-000000000000 \
    --resource-group uit-go-dev-rg \
    --location eastus \
    --acr-name uitgodevacr \
    --key-vault uitgodevkv
EOF
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

write_generated_vars() {
  echo "==> Writing ${GENERATED_VARS}"
  cat >"${GENERATED_VARS}" <<EOF
{
  "resource_group_name": "${RESOURCE_GROUP}",
  "environment": "${ENVIRONMENT}",
  "acr_name": "${ACR_NAME}",
  "key_vault_name": "${KEY_VAULT_NAME}",
  "container_apps": {
    "api-gateway": {
      "app_name": "api-gateway",
      "image_repository": "api-gateway",
      "image_tag": "${TAG}",
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": true,
        "target_port": 8080
      },
      "environment_variables": {
        "PORT": "8080",
        "ENVIRONMENT": "${ENVIRONMENT}",
        "OTEL_EXPORTER": "otlp",
        "OTEL_COLLECTOR_ENDPOINT": "jaeger.internal:4317"
      }
    },
    "authentication-service": {
      "app_name": "authentication-service",
      "image_repository": "authentication-service",
      "image_tag": "${TAG}",
      "cpu": 0.5,
      "memory": "1Gi",
      "environment_variables": {
        "DSN": "host=postgres.internal port=5432 user=postgres password=password dbname=users sslmode=disable timezone=UTC connect_timeout=5",
        "JWT_SECRET": "your-secret-key-change-in-production",
        "JWT_EXPIRY": "24h",
        "REFRESH_TOKEN_EXPIRY": "168h",
        "OTEL_EXPORTER": "otlp",
        "OTEL_COLLECTOR_ENDPOINT": "jaeger.internal:4317"
      }
    },
    "location-service": {
      "app_name": "location-service",
      "image_repository": "location-service",
      "image_tag": "${TAG}",
      "cpu": 0.5,
      "memory": "1Gi",
      "environment_variables": {
        "REDIS_HOST": "redis.internal",
        "REDIS_PORT": "6379",
        "REDIS_PASSWORD": "redispassword",
        "REDIS_DB": "0",
        "REDIS_TIME_TO_LIVE": "3600",
        "OTEL_EXPORTER": "otlp",
        "OTEL_COLLECTOR_ENDPOINT": "jaeger.internal:4317"
      }
    },
    "logger-service": {
      "app_name": "logger-service",
      "image_repository": "logger-service",
      "image_tag": "${TAG}",
      "cpu": 0.5,
      "memory": "1Gi",
      "environment_variables": {
        "MONGO_URL": "mongodb://admin:password@mongo.internal:27017",
        "OTEL_EXPORTER": "otlp",
        "OTEL_COLLECTOR_ENDPOINT": "jaeger.internal:4317"
      }
    },
    "trip-service": {
      "app_name": "trip-service",
      "image_repository": "trip-service",
      "image_tag": "${TAG}",
      "cpu": 0.75,
      "memory": "1.5Gi",
      "environment_variables": {
        "DSN": "host=postgres-trip.internal port=5432 user=postgres password=password dbname=trips sslmode=disable timezone=UTC connect_timeout=5",
        "HERE_ID": "${HERE_ID_VALUE}",
        "HERE_SECRET": "${HERE_SECRET_VALUE}",
        "OTEL_EXPORTER": "otlp",
        "OTEL_COLLECTOR_ENDPOINT": "jaeger.internal:4317"
      }
    },
    "postgres": {
      "app_name": "postgres",
      "image_uri": "docker.io/library/postgres:16-alpine",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": false,
        "target_port": 5432,
        "transport": "tcp"
      },
      "environment_variables": {
        "POSTGRES_USER": "postgres",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DB": "users"
      }
    },
    "postgres_trip": {
      "app_name": "postgres-trip",
      "image_uri": "docker.io/library/postgres:16-alpine",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": false,
        "target_port": 5432,
        "transport": "tcp"
      },
      "environment_variables": {
        "POSTGRES_USER": "postgres",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DB": "trips"
      }
    },
    "mongo": {
      "app_name": "mongo",
      "image_uri": "docker.io/library/mongo:7-jammy",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": false,
        "target_port": 27017,
        "transport": "tcp"
      },
      "environment_variables": {
        "MONGO_INITDB_DATABASE": "logs",
        "MONGO_INITDB_ROOT_USERNAME": "admin",
        "MONGO_INITDB_ROOT_PASSWORD": "password"
      }
    },
    "redis": {
      "app_name": "redis",
      "image_uri": "docker.io/library/redis:7-alpine",
      "use_acr": false,
      "cpu": 0.25,
      "memory": "0.5Gi",
      "command": [
        "redis-server",
        "--requirepass",
        "redispassword"
      ],
      "ingress": {
        "external": false,
        "target_port": 6379,
        "transport": "tcp"
      },
      "environment_variables": {
        "REDIS_PASSWORD": "redispassword"
      }
    },
    "rabbitmq": {
      "app_name": "rabbitmq",
      "image_uri": "docker.io/library/rabbitmq:3.13-management-alpine",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": false,
        "target_port": 5672,
        "transport": "tcp"
      },
      "environment_variables": {
        "RABBITMQ_DEFAULT_USER": "guest",
        "RABBITMQ_DEFAULT_PASS": "guest"
      }
    },
    "jaeger": {
      "app_name": "jaeger",
      "image_uri": "docker.io/jaegertracing/jaeger:2.2.0",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": false,
        "target_port": 4317,
        "transport": "tcp"
      },
      "environment_variables": {
        "COLLECTOR_OTLP_ENABLED": "true"
      }
    },
    "prometheus": {
      "app_name": "prometheus",
      "image_uri": "docker.io/prom/prometheus:v3.2.0",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "args": [
        "--config.file=/etc/prometheus/prometheus.yml",
        "--storage.tsdb.path=/prometheus"
      ],
      "ingress": {
        "external": false,
        "target_port": 9090
      }
    },
    "grafana": {
      "app_name": "grafana",
      "image_uri": "docker.io/grafana/grafana:11.5.0",
      "use_acr": false,
      "cpu": 0.5,
      "memory": "1Gi",
      "ingress": {
        "external": true,
        "target_port": 3000
      },
      "environment_variables": {
        "GF_SECURITY_ADMIN_PASSWORD": "admin",
        "GF_USERS_ALLOW_SIGN_UP": "false"
      }
    }
  }
}
EOF
}

# Default values
ENVIRONMENT="dev"
TAG="$(git -C "${REPO_ROOT}" rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)"
TFVARS_FILE=""
SKIP_BUILD="false"
SKIP_APPLY="false"
SUBSCRIPTION=""
RESOURCE_GROUP=""
LOCATION=""
ACR_NAME=""
KEY_VAULT_NAME=""
HERE_ID_VALUE="${HERE_ID:-}"
HERE_SECRET_VALUE="${HERE_SECRET:-}"

if [[ -z "${HERE_ID_VALUE}" || -z "${HERE_SECRET_VALUE}" ]]; then
  echo "Warning: HERE_ID or HERE_SECRET not set; trip-service will be deployed without HERE credentials." >&2
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --subscription)
      SUBSCRIPTION="$2"
      shift 2
      ;;
    --resource-group)
      RESOURCE_GROUP="$2"
      shift 2
      ;;
    --location)
      LOCATION="$2"
      shift 2
      ;;
    --acr-name)
      ACR_NAME="$2"
      shift 2
      ;;
    --key-vault)
      KEY_VAULT_NAME="$2"
      shift 2
      ;;
    --env)
      ENVIRONMENT="$2"
      shift 2
      ;;
    --tag)
      TAG="$2"
      shift 2
      ;;
    --tfvars)
      TFVARS_FILE="$2"
      shift 2
      ;;
    --skip-build)
      SKIP_BUILD="true"
      shift
      ;;
    --skip-apply)
      SKIP_APPLY="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${SUBSCRIPTION}" || -z "${RESOURCE_GROUP}" || -z "${LOCATION}" || -z "${ACR_NAME}" || -z "${KEY_VAULT_NAME}" ]]; then
  echo "Error: subscription, resource group, location, ACR name, and Key Vault name are required." >&2
  usage
  exit 1
fi

REQUIRED_CMDS=(az tofu git)
if [[ "${SKIP_BUILD}" != "true" ]]; then
  REQUIRED_CMDS+=(podman)
fi

for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command_exists "$cmd"; then
    echo "Error: required command '$cmd' not found in PATH." >&2
    exit 1
  fi
done

az account set --subscription "${SUBSCRIPTION}"
az group create --name "${RESOURCE_GROUP}" --location "${LOCATION}" >/dev/null

ARM_SUBSCRIPTION_ID_VALUE="$(az account show --query id -o tsv)"
ARM_TENANT_ID_VALUE="$(az account show --query tenantId -o tsv)"
export ARM_SUBSCRIPTION_ID="${ARM_SUBSCRIPTION_ID_VALUE}"
export ARM_TENANT_ID="${ARM_TENANT_ID_VALUE}"
export ARM_USE_AZUREAD="true"

REGISTRY_SERVER="${ACR_NAME}.azurecr.io"
write_generated_vars

TF_VAR_ARGS=()
if [[ -n "${TFVARS_FILE}" ]]; then
  TF_VAR_ARGS+=("-var-file=${TFVARS_FILE}")
fi

pushd "${TF_DIR}" >/dev/null

echo "==> Formatting OpenTofu configuration"
tofu fmt

PLAN_FILE="tfplan"

echo "==> Initializing OpenTofu"
tofu init

echo "==> Validating OpenTofu configuration"
tofu validate

if ! az acr show --name "${ACR_NAME}" >/dev/null 2>&1; then
  echo "==> Provisioning container registry prerequisites"
  BOOTSTRAP_PLAN="bootstrap.tfplan"
  tofu plan "${TF_VAR_ARGS[@]}" -target=azurerm_container_registry.this -out "${BOOTSTRAP_PLAN}"
  tofu apply "${BOOTSTRAP_PLAN}"
  rm -f "${BOOTSTRAP_PLAN}"
fi

if [[ "${SKIP_BUILD}" != "true" ]]; then
  az acr login --name "${ACR_NAME}"

  for service in "${SERVICES[@]}"; do
    context_dir="${REPO_ROOT}"
    service_dir="${REPO_ROOT}/${SERVICE_CONTEXTS[$service]}"
    dockerfile_path="${service_dir}/${SERVICE_DOCKERFILES[$service]}"
    image_tag="${REGISTRY_SERVER}/${service}:${TAG}"

    echo
    echo "==> Building ${service}"
    podman build \
      --file "${dockerfile_path}" \
      --tag "${image_tag}" \
      "${context_dir}"

    echo "==> Pushing ${service}"
    podman push "${image_tag}"
  done
fi

echo "==> Planning infrastructure changes"
tofu plan "${TF_VAR_ARGS[@]}" -out "${PLAN_FILE}"

if [[ "${SKIP_APPLY}" != "true" ]]; then
  echo "==> Applying infrastructure changes"
  tofu apply "${PLAN_FILE}"
else
  echo "Skipping apply step (--skip-apply supplied)"
fi

popd >/dev/null

echo "\nDeployment workflow completed. Review the outputs above for resource endpoints."
