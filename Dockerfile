FROM golang:1.24-alpine as go-build
RUN apk add --no-cache \
  git \
  ca-certificates \
  gcc \
  musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN ls -la && go mod download
COPY ./ ./
RUN go build -o ./main main.go

FROM node:18 as node-build
WORKDIR /app
COPY ./front/ ./
RUN npm install && npm run build


FROM golang:1.24-alpine
WORKDIR /home/app
COPY --from=go-build /app/main /home/app/
COPY --from=node-build /app/build /home/app/
CMD ["./main"]