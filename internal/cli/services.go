package cli

import "strings"

type service struct {
	Name        string
	App         string
	Detect      []string
	Requires    []string
	Deps        []string
	Env         string
	CfgField    string
	CfgType     string
	CfgValue    string
	CfgValidate string
	Imports     []string
	Middleware  string
	Conn        string
	Compose     string
	Volume      string
	Clean       *cleanTarget
	DockerEnv   string
	Healthy     bool
}

type cleanTarget struct {
	Name   string
	Script string
}

var services = map[string]service{
	"postgres": {
		Name: "postgres",
		Deps: []string{"github.com/lib/pq", "github.com/golang-migrate/migrate/v4"},
		Env: `# Database
POSTGRES_URI=postgres://{%.Name%}_app:{%.Name%}_app@localhost:5432/{%.Name%}?sslmode=disable
POSTGRES_MIGRATION_URI=postgres://postgres:postgres@localhost:5432/{%.Name%}?sslmode=disable
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_APP_USER={%.Name%}_app
POSTGRES_APP_PASSWORD={%.Name%}_app
POSTGRES_DB={%.Name%}
`,
		CfgField: "PostgreSQL PostgreSQL",
		CfgType: `type PostgreSQL struct {
	Uri          string
	MigrationUri string
}
`,
		CfgValue: `PostgreSQL: PostgreSQL{
	Uri:          os.Getenv("POSTGRES_URI"),
	MigrationUri: os.Getenv("POSTGRES_MIGRATION_URI"),
},`,
		CfgValidate: `if cfg.PostgreSQL.Uri == "" {
	erros = append(erros, "POSTGRES_URI")
}

if cfg.PostgreSQL.MigrationUri == "" {
	erros = append(erros, "POSTGRES_MIGRATION_URI")
}
`,
		Imports: []string{`"{%.Module%}/pkg/postgres"`, `"{%.Module%}/pkg/migration"`},
		Conn: `if err := migration.Migrate(ctx, cfg.PostgreSQL); err != nil {
	errorx.Fatal(err)
}

pgdb, err := postgres.NewPostgresConn(ctx, cfg.PostgreSQL)
if err != nil {
	errorx.Fatal(err)
}
defer pgdb.Close()

`,
		Compose: `postgres:
  image: postgres:18-alpine
  container_name: {%.Name%}-postgres
  restart: unless-stopped
  env_file:
    - ../../.env
  ports:
    - "${POSTGRES_PORT:-5432}:5432"
  volumes:
    - postgres_data:/var/lib/postgresql
    - ./postgres:/docker-entrypoint-initdb.d:ro
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
    interval: 5s
    timeout: 5s
    retries: 5
`,
		Volume:    "postgres_data:",
		DockerEnv: "POSTGRES_URI: postgres://${POSTGRES_APP_USER}:${POSTGRES_APP_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable\nPOSTGRES_MIGRATION_URI: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable",
		Healthy:   true,
		Clean: &cleanTarget{
			Name: "database",
			Script: `if [ "$(confirm)" != "true" ]; then \
	echo "limpar o database apaga todos os dados: rode com confirm=true (ex: make sudo-clean database confirm=true)"; \
	exit 1; \
fi; \
echo "limpando banco de dados..."; \
printf '%s\n' "SELECT format('TRUNCATE TABLE %I RESTART IDENTITY CASCADE', tablename) FROM pg_tables WHERE schemaname = current_schema() AND tablename <> 'schema_migrations' \gexec" | \
sudo docker compose -f infra/compose/docker-compose.yaml --env-file .env exec -T postgres \
	sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'; \`,
		},
	},
	"redis": {
		Name: "redis",
		Deps: []string{"github.com/redis/go-redis/v9"},
		Env: `# Redis
REDIS_URI=redis://:redis@localhost:6379/0
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis
`,
		CfgField: "Redis Redis",
		CfgType: `type Redis struct {
	Uri string
}
`,
		CfgValue: `Redis: Redis{
	Uri: os.Getenv("REDIS_URI"),
},`,
		CfgValidate: `if cfg.Redis.Uri == "" {
	erros = append(erros, "REDIS_URI")
}
`,
		Imports: []string{`"{%.Module%}/pkg/redis"`},
		Conn: `rdb, err := redis.NewRedisConn(ctx, cfg.Redis)
if err != nil {
	errorx.Fatal(err)
}
defer rdb.Close()

`,
		Compose: `redis:
  image: redis:8-alpine
  container_name: {%.Name%}-redis
  restart: unless-stopped
  env_file:
    - ../../.env
  command: ["redis-server", "--requirepass", "${REDIS_PASSWORD}"]
  ports:
    - "${REDIS_PORT}:6379"
  volumes:
    - redis_data:/data
  healthcheck:
    test: ["CMD-SHELL", "redis-cli -a $${REDIS_PASSWORD} ping | grep -q PONG"]
    interval: 5s
    timeout: 5s
    retries: 5
`,
		Volume:    "redis_data:",
		DockerEnv: "REDIS_URI: redis://:${REDIS_PASSWORD}@redis:6379/0",
		Healthy:   true,
		Clean: &cleanTarget{
			Name: "cache",
			Script: `echo "limpando cache..."; \
sudo docker compose -f infra/compose/docker-compose.yaml --env-file .env exec -T redis \
	sh -c 'redis-cli -a "$$REDIS_PASSWORD" FLUSHALL'; \`,
		},
	},
	"rabbitmq": {
		Name: "rabbitmq",
		Deps: []string{"github.com/rabbitmq/amqp091-go"},
		Env: `# RabbitMQ
RABBITMQ_URI=amqp://rabbitmq:rabbitmq@localhost:5672/
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=rabbitmq
RABBITMQ_PASSWORD=rabbitmq
RABBITMQ_MANAGEMENT_PORT=15672
`,
		CfgField: "RabbitMQ RabbitMQ",
		CfgType: `type RabbitMQ struct {
	Uri string
}
`,
		CfgValue: `RabbitMQ: RabbitMQ{
	Uri: os.Getenv("RABBITMQ_URI"),
},`,
		CfgValidate: `if cfg.RabbitMQ.Uri == "" {
	erros = append(erros, "RABBITMQ_URI")
}
`,
		Imports: []string{`"{%.Module%}/pkg/rabbitmq"`},
		Conn: `rmq, err := rabbitmq.NewRabbitMQConn(ctx, cfg.RabbitMQ)
if err != nil {
	errorx.Fatal(err)
}
defer rmq.Close()

`,
		Compose: `rabbitmq:
  image: rabbitmq:4-management-alpine
  container_name: {%.Name%}-rabbitmq
  restart: unless-stopped
  environment:
    RABBITMQ_DEFAULT_USER: ${RABBITMQ_USER}
    RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASSWORD}
  ports:
    - "${RABBITMQ_PORT:-5672}:5672"
    - "${RABBITMQ_MANAGEMENT_PORT:-15672}:15672"
  volumes:
    - rabbitmq_data:/var/lib/rabbitmq
  healthcheck:
    test: ["CMD", "rabbitmq-diagnostics", "-q", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5
`,
		Volume:    "rabbitmq_data:",
		DockerEnv: "RABBITMQ_URI: amqp://${RABBITMQ_USER}:${RABBITMQ_PASSWORD}@rabbitmq:5672/",
		Healthy:   true,
	},
	"resend": {
		Name: "resend",
		Deps: []string{"github.com/resend/resend-go/v3"},
		Env: `# Resend
RESEND_API_KEY=
RESEND_FROM=
`,
		CfgField: "Resend Resend",
		CfgType: `type Resend struct {
	ApiKey string
	From   string
}
`,
		CfgValue: `Resend: Resend{
	ApiKey: os.Getenv("RESEND_API_KEY"),
	From:   os.Getenv("RESEND_FROM"),
},`,
		CfgValidate: `if cfg.Resend.ApiKey == "" {
	erros = append(erros, "RESEND_API_KEY")
}

if cfg.Resend.From == "" {
	erros = append(erros, "RESEND_FROM")
}
`,
	},
	"stripe": {
		Name: "stripe",
		Deps: []string{"github.com/stripe/stripe-go/v85"},
		Env: `# Stripe
STRIPE_SECRET_KEY=
STRIPE_PUBLIC_KEY=
STRIPE_WEBHOOK_SECRET=
`,
		CfgField: "Stripe Stripe",
		CfgType: `type Stripe struct {
	SecretKey     string
	PublicKey     string
	WebhookSecret string
}
`,
		CfgValue: `Stripe: Stripe{
	SecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
	PublicKey:     os.Getenv("STRIPE_PUBLIC_KEY"),
	WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
},`,
		CfgValidate: `if cfg.Stripe.SecretKey == "" {
	erros = append(erros, "STRIPE_SECRET_KEY")
}

if cfg.Stripe.PublicKey == "" {
	erros = append(erros, "STRIPE_PUBLIC_KEY")
}
`,
	},
	"prometheus": {
		Name:    "prometheus",
		Detect:  []string{"infra", "compose", "prometheus"},
		Deps:    []string{"github.com/zsais/go-gin-prometheus"},
		Imports: []string{`ginprometheus "github.com/zsais/go-gin-prometheus"`},
		Middleware: `prometheus := ginprometheus.NewWithConfig(ginprometheus.Config{Subsystem: "gin"})
prometheus.Use(api)
`,
		Env: `# Prometheus
PROMETHEUS_PORT=9090
`,
		Compose: `prometheus:
  image: prom/prometheus:latest
  container_name: {%.Name%}-prometheus
  restart: unless-stopped
  volumes:
    - ./prometheus:/etc/prometheus:ro
    - prometheus_data:/prometheus
  extra_hosts:
    - "host.docker.internal:host-gateway"
  ports:
    - "${PROMETHEUS_PORT}:9090"
`,
		Volume: "prometheus_data:",
	},
	"grafana": {
		Name:     "grafana",
		Detect:   []string{"infra", "compose", "grafana"},
		Requires: []string{"prometheus"},
		Env: `# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin
GRAFANA_PORT=3000
`,
		Compose: `grafana:
  image: grafana/grafana:latest
  container_name: {%.Name%}-grafana
  restart: unless-stopped
  depends_on:
    - prometheus
  volumes:
    - ./grafana:/etc/grafana/provisioning:ro
    - grafana_data:/var/lib/grafana
  environment:
    GF_SECURITY_ADMIN_USER: ${GRAFANA_USER}
    GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
  ports:
    - "${GRAFANA_PORT}:3000"
`,
		Volume: "grafana_data:",
	},
}

var dockerService = service{
	Name:   "docker",
	Detect: []string{"Dockerfile"},
	Compose: `api:
  build:
    context: ../..
  container_name: {%.Name%}-api
  restart: unless-stopped
  profiles: ["app"]
  env_file:
    - ../../.env
  environment:
    GIN_MODE: release
  ports:
    - "${APPLICATION_PORT}:${APPLICATION_PORT}"
`,
}

var viperService = service{
	Name:    "viper",
	App:     "cli",
	Detect:  []string{"internal", "commands", "config", "viper.go"},
	Deps:    []string{"github.com/spf13/viper"},
	Imports: []string{`"{%.Module%}/internal/commands/config"`},
	Conn: `config.NewViper()
`,
}

func init() {
	services["docker"] = dockerService
	services["viper"] = viperService
}

// serviceOrder keeps the makefile clean block stable.
var serviceOrder = []string{"postgres", "redis", "rabbitmq", "resend", "stripe", "prometheus", "grafana", "docker", "viper"}

// app is where the service lives: api services by default, viper in the cli.
func (s service) app() string {
	if s.App == "" {
		return "api"
	}
	return s.App
}

func (s service) mainPath() []string {
	return []string{"cmd", s.app(), "main.go"}
}

func (s service) composeKey() string {
	key, _, _ := strings.Cut(s.Compose, ":")
	return key
}

func (s service) detectPath() []string {
	if s.Detect != nil {
		return s.Detect
	}
	return []string{"pkg", s.Name}
}
