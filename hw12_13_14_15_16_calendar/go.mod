module github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar

go 1.23.0

require (
	github.com/getkin/kin-openapi v0.124.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.5.0
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jackc/pgx/v5 v5.5.2
	github.com/oapi-codegen/runtime v1.1.1
	github.com/pressly/goose/v3 v3.19.0
	github.com/prometheus/client_golang v1.19.1
	github.com/segmentio/kafka-go v0.4.51
	github.com/stretchr/testify v1.9.0
	github.com/swaggo/http-swagger/v2 v2.0.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.22.8 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	github.com/swaggo/swag v1.8.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.20.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

// goose с duckdb драйвером вызывает компиляцию duckdb, которая не работает
// на Windows. Используем build constraints для исключения duckdb при сборке goose.
//go:build !duckdb
// +build !duckdb
