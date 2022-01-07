# Prometheus metrics for Lambda HTTP Gateway

A worked example of using Prometheus to scrape metrics from an Lambda HTTP Gateway instance.

## Prerequisites

- Docker
- (Optional) Docker Compose

## Docker Compose

You can use the Docker Compose file in this directory to quickly stand up Prometheus and Grafana.

## Plain Docker

If you don't want to use Docker Compose, you can start the individual containers.

Start Prometheus using the configuration in this directory:

	docker run --rm -it -p9090:9090 -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

(Optional) start Grafana:

    docker run --rm -it -p3000:3000 -v $PWD/grafana.yml:/etc/grafana/provisioning/datasources/lambdahttpgw.yml grafana/grafana
