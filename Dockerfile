# syntax=docker/dockerfile:1
# Multi-stage build for any yw-mall Go service.
# Usage: docker build --build-arg SERVICE=mall-user-rpc -t mall-user-rpc .
FROM golang:1.26-alpine AS builder
WORKDIR /workspace

COPY mall-common                  ./mall-common
COPY mall-activity-async-worker   ./mall-activity-async-worker
COPY mall-activity-rpc            ./mall-activity-rpc
COPY mall-api                     ./mall-api
COPY mall-cart-rpc                ./mall-cart-rpc
COPY mall-logistics-rpc           ./mall-logistics-rpc
COPY mall-order-rpc               ./mall-order-rpc
COPY mall-payment-rpc             ./mall-payment-rpc
COPY mall-product-rpc             ./mall-product-rpc
COPY mall-review-rpc              ./mall-review-rpc
COPY mall-reward-rpc              ./mall-reward-rpc
COPY mall-risk-rpc                ./mall-risk-rpc
COPY mall-rule-rpc                ./mall-rule-rpc
COPY mall-shop-rpc                ./mall-shop-rpc
COPY mall-user-rpc                ./mall-user-rpc
COPY mall-workflow-rpc            ./mall-workflow-rpc

ARG SERVICE
RUN cd /workspace/${SERVICE} && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/server . && \
    cp -r etc /out/etc

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /out/server  ./server
COPY --from=builder /out/etc     ./etc
COPY docker-entrypoint.sh        ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]
