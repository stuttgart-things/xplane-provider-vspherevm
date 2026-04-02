# Stage 1: Build the provider binary
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /provider ./cmd/provider

# Stage 2: Runtime image with Terraform CLI and provider plugin
FROM alpine:3.23.2

RUN apk --no-cache add ca-certificates bash

ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Copy provider binary from builder stage
COPY --from=builder /provider /usr/local/bin/provider

ENV USER_ID=65532

# Setup Terraform environment

## Provider-dependent configuration
ARG TERRAFORM_VERSION=1.5.7
ARG TERRAFORM_PROVIDER_SOURCE=vmware/vsphere
ARG TERRAFORM_PROVIDER_VERSION=2.15.0
ARG TERRAFORM_PROVIDER_DOWNLOAD_NAME=terraform-provider-vsphere
ARG TERRAFORM_NATIVE_PROVIDER_BINARY=terraform-provider-vsphere_v2.15.0_x5
ARG TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX=https://releases.hashicorp.com/terraform-provider-vsphere

## End of - Provider-dependent configuration

ENV PLUGIN_DIR=/terraform/provider-mirror/registry.terraform.io/${TERRAFORM_PROVIDER_SOURCE}/${TERRAFORM_PROVIDER_VERSION}/${TARGETOS}_${TARGETARCH}
ENV TF_CLI_CONFIG_FILE=/terraform/.terraformrc
ENV TF_FORK=0

RUN mkdir -p ${PLUGIN_DIR}

ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp
ADD ${TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX}/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp
ADD terraformrc.hcl ${TF_CLI_CONFIG_FILE}

RUN unzip /tmp/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip -d /usr/local/bin \
  && chmod +x /usr/local/bin/terraform \
  && rm /tmp/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip \
  && unzip /tmp/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip -d ${PLUGIN_DIR} \
  && chmod +x ${PLUGIN_DIR}/* \
  && rm /tmp/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip \
  && chown -R ${USER_ID}:${USER_ID} /terraform
# End of - Setup Terraform environment

# Provider controller needs these environment variable at runtime
ENV TERRAFORM_VERSION=${TERRAFORM_VERSION}
ENV TERRAFORM_PROVIDER_SOURCE=${TERRAFORM_PROVIDER_SOURCE}
ENV TERRAFORM_PROVIDER_VERSION=${TERRAFORM_PROVIDER_VERSION}
ENV TERRAFORM_NATIVE_PROVIDER_PATH=${PLUGIN_DIR}/${TERRAFORM_NATIVE_PROVIDER_BINARY}

USER ${USER_ID}
EXPOSE 8080

ENTRYPOINT ["provider"]
