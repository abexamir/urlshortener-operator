# URL Shortener Operator

A Kubernetes operator that manages URL shortening services within a cluster. It creates and manages short URLs through CustomResources and provides a redirection service.

## Features

- Custom Resource Definition (CRD) for managing short URLs
- Automatic short path generation using SHA-256 hashing
- Redis-backed storage for URL mappings
- Click tracking for each short URL
- HTTP redirection server (port 8082)
- Metrics endpoint (port 8080)
- Health probes (port 8081)

## Architecture

The operator consists of three main components:

1. **Controller**: Watches for ShortURL resources and manages their lifecycle
2. **Redis Service**: Handles URL storage and click tracking
3. **HTTP Server**: Handles redirections for short URLs

## Prerequisites

- Go version v1.23.0+
- Docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

## Installation

### Using pre-built images

1. Install the CRDs and the Operator:
```sh
kubectl apply -f https://raw.githubusercontent.com/abexamir/urlshortener-operator/refs/heads/main/dist/install.yaml
```


### Using the Helm chart

1. Clone the repository
```sh
git clone git@github.com:abexamir/urlshortener-operator.git && cd urlshortener-operator # clone the repository first
```
2. Install helm chart 
```sh
helm install urlshortener-operator ./dist/chart --namespace urlshortener-operator --create-namespace
```


### Building from source

1. Clone the repository:
```sh
git clone git@github.com:abexamir/urlshortener-operator.git && cd urlshortener-operator # clone the repository first
```

2. Build and push the operator image:
```sh
make docker-build docker-push IMG=<your-registry>/url-shortener-operator:tag
```

3. Deploy to cluster:
```sh
make deploy IMG=<your-registry>/url-shortener-operator:tag
```

## Usage

1. Create a ShortURL resource:
```yaml
apiVersion: urlshortener.tapsi.ir/v1
kind: ShortURL
metadata:
  name: example-url
spec:
  targetURL: "https://example.com"
```

2. Apply the resource:
```sh
kubectl apply -f shorturl.yaml
```

3. Check the status:
```sh
kubectl get shorturl example-url -o yaml
```

The status section will contain:
- `shortPath`: The generated short path
- `clickCount`: Number of times the URL has been accessed  
Please note that the clickCount field get eventually consistent and doesn't get updated instantly (To put less pressure on the API Server)

4. Access the shortened URL:
```sh
http://<operator-service>/<shortPath> # e.g. http://<operator-service>/abc
# The pod will listen on port 8082 and there's a service in front of it with port 80, if you have ingress controller installed, you can enable the ingress resource by setting `ingress.enable` to true and set `ingress.host` to your ingress host in the helm chart values. For test purposes, just use simple port-forwarding.
```

5. If you update the `targetURL` field, the short path will be updated immediately and the click count will be reset.

## Development

### Local Development

1. Install dependencies:
```sh
go mod download
```

2. Run the operator locally:
```sh
make run
```


### Building

```sh
# Build binary
make build

# Build Docker image
make docker-build IMG=<your-registry>/url-shortener-operator:tag
```

## Monitoring

The operator exposes metrics in Prometheus format at `:8080/metrics` (or `:8443/metrics` if secure metrics are enabled).

Health endpoints:
- Liveness: `:8081/healthz`
- Readiness: `:8081/readyz`

## Cleanup

1. Remove all ShortURL resources:
```sh
kubectl delete shorturls --all
```

2. Uninstall the operator:
```sh
make undeploy
```

3. Remove the CRDs:
```sh
make uninstall
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
