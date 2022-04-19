docker-compose up -d influxdb grafana
docker-compose run k6 run --vus 500 --duration 60s /scripts/stress-test.js