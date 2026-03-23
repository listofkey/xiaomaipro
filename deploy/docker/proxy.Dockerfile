# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS front-builder

ARG PNPM_VERSION=10.25.0

WORKDIR /src/front

RUN npm install -g pnpm@${PNPM_VERSION}

COPY front/package.json front/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY front ./

ARG FRONT_VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${FRONT_VITE_API_BASE_URL}

RUN pnpm build

FROM node:22-alpine AS admin-builder

ARG PNPM_VERSION=10.25.0

WORKDIR /src/web

RUN npm install -g pnpm@${PNPM_VERSION}

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web ./

ARG ADMIN_BASE_PATH=/admin/

RUN pnpm build -- --base=${ADMIN_BASE_PATH}

FROM nginx:1.28-alpine

COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=front-builder /src/front/dist /usr/share/nginx/html/front
COPY --from=admin-builder /src/web/dist /usr/share/nginx/html/admin
