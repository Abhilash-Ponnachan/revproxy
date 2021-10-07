FROM golang:1.17-alpine AS build
WORKDIR /revproxy
COPY go.mod *go.sum *.go ./
RUN CGO_ENABLED=0 go build -o ./proxy

FROM alpine
WORKDIR /revproxy
COPY *.json ./
COPY --from=build /revproxy/proxy ./
ENV PORT 9090
EXPOSE ${PORT}
#CMD tail -f /dev/null
CMD [ "./proxy" ]
