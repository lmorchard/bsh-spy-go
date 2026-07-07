# syntax=docker/dockerfile:1
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bsh-spy-go .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bsh-spy-go /bsh-spy-go
VOLUME ["/data"]
ENTRYPOINT ["/bsh-spy-go"]
CMD ["run"]
