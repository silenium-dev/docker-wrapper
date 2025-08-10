package resources

import _ "embed"

//go:embed logging_stack/config.alloy
var alloyConfig string

//go:embed logging_stack/main.alloy
var alloyMainConfig string

//go:embed logging_stack/docker-compose.yaml
var composeString string

//go:embed logging_stack/loki.yaml
var lokiConfig string

//go:embed logging_stack/grafana-loki.yaml
var grafanaLokiDatasource string
