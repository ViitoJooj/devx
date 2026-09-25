# devx
Binário único que gera e evolui projetos Go no padrão do noxacloud.
Os comandos seguem o padrão do gh: `devx <substantivo> <verbo>`.

## Instalação
```sh
go build -o $(go env GOPATH)/bin/devx ./cmd
```

## Comandos
```sh
devx init <project>                  # api (gin, padrão noxacloud) em cmd/api (padrão)
devx init <project> --api            # o mesmo, explícito
devx init <project> --cli            # cli (cobra) em cmd/cli, comandos em internal/commands; config.yaml via `service add viper`
devx init <project> --api --cli      # os dois no mesmo módulo, dividindo pkg/
                                     # --module github.com/ViitoJooj/<project> muda o módulo

devx app add api|cli                 # adiciona a outra parte depois
devx app rm api|cli --sudo [-y]      # remove a parte inteira (migrations ficam; não remove o último app)
devx app ls [--json]

devx command add get-jobs            # internal/commands/get_jobs.go registrado no root
devx command rm get-jobs [-y]
devx command ls [--json]

devx service add postgres redis rabbitmq resend stripe prometheus grafana docker   # api
devx service add viper                                                             # cli (config.yaml)
devx service rm redis [-y]
devx service ls [--json]

devx crud add user [--db postgres|memory] [--no-migrations]
devx crud db user postgres|memory [-y]   # troca o repositório do crud
devx crud rm user [-y]              # migrations ficam
devx crud rm server [-y]            # remove o container da api (/v1/ping, /v1/openapi.yaml e /docs)
devx crud grant user delete         # libera DELETE na tabela para a api (gera migration)
devx crud revoke audit update       # bloqueia UPDATE (vários de uma vez: revoke setting insert update)
devx crud ls [--json]               # rota, banco, acesso (ACCESS) e openapi de cada crud

devx worker add cleanup [--interval 30m]
devx worker rm cleanup [-y]
devx worker ls [--json]

devx view add sveltekit|react|next|tauri|tauri-svelte|tauri-react [--name admin]
devx view rm sveltekit [-y]
devx view ls [--json]
```

- Nada é criado antes de ser usado: `migrations/` nasce no primeiro `crud add`,
  `internal/workers` e `workers.New(...)` no primeiro `worker add`, `rootCmd.AddCommand(...)` no primeiro `command add`,
  o compose no primeiro serviço com container; e somem quando o último é removido.
- `crud`, `worker` e `view` exigem a api; `command` exige a cli; cada `service` exige o app dele
  (`devx service ls` mostra a coluna APP e "needs api/cli").
- `ls` também responde por `list` e `rm` por `remove`.
- `rm` pede confirmação; `-y`/`--yes` pula (obrigatório fora de um terminal, ex: CI).
- `service add grafana` adiciona o prometheus; `service add prometheus` injeta o middleware do gin e expõe `/metrics`.
- `service add resend` / `stripe` não criam conexão no `main.go`: use `resend.NewResendClient(cfg.Resend)` / `stripe.NewStripeClient(cfg.Stripe)` onde precisar.
- `service add docker`: gera `Dockerfile` e `.dockerignore` e coloca a API no compose no profile `app`.
  `make run-dev` continua subindo só a infra + air; `make run-docker` sobe tudo em container.
- `service rm` recusa remover o que ainda está em uso (postgres com cruds em postgres, prometheus com grafana).
- Postgres com privilégio mínimo: a API conecta com `POSTGRES_URI` (usuário `<projeto>_app`: só SELECT/INSERT/UPDATE,
  sem DELETE/TRUNCATE/DROP/ALTER/CREATE e sem acesso a `schema_migrations`); as migrations rodam com
  `POSTGRES_MIGRATION_URI` (dono do banco). O usuário é criado por `infra/compose/postgres/init.sh` na primeira
  subida do container; num banco que já existia, apague o volume (`docker compose ... down -v`) e suba de novo.
  `make sudo-clean database confirm=true` limpa os dados (TRUNCATE) mantendo `schema_migrations`.
- `crud grant|revoke` ajusta isso por tabela (`select`, `insert`, `update`, `delete`) sempre por migration
  (`000N_Audit_revoke_update.up/down.sql`), então vale em qualquer ambiente e o `down` desfaz. O padrão de toda
  tabela nova é SELECT/INSERT/UPDATE; a coluna ACCESS do `crud ls` reconstrói o acesso atual lendo as migrations
  (`-` para crud em memória, `unknown` se nenhuma migration cria a tabela).
  As rotas do crud não mudam: a operação bloqueada responde o erro de permissão do banco.
- Banco do crud: com postgres no projeto, `crud add` usa postgres. Sem banco, ele pergunta
  (adicionar postgres / repositório em memória / cancelar); fora de um terminal exige `--db`.
  O repositório em memória (map + mutex) implementa o mesmo `contracts.IRepository`: tudo funciona sem banco,
  mas os dados somem ao reiniciar. `crud db` troca depois (repository.go, initializr.go e a chamada no main.go),
  criando a migration se ainda não existir. `--no-migrations` não gera a migration (tabela já existe).
- `view add` cria o frontend em `www/<nome>` com o CLI oficial via bun (`sv`, `create-vite`, `create-next-app`,
  `create-tauri-app`; precisa de internet). `tauri` sozinho é HTML + CSS puro.
  Cada view ganha uma porta própria (sveltekit 5173, react 5174+, next 3001; tauri-svelte/react usam 1420 fixo),
  a origem vai para `APPLICATION_FRONTEND` no `.env` (lista separada por vírgula, com `tauri://localhost` para as views tauri)
  e a API ganha CORS (`pkg/httpx/cors.go`) na primeira view. `make run-dev` sobe a API e todas as views em paralelo;
  `make view-<nome>` sobe uma só.
- OpenAPI: `docs/openapi/openapi.yaml` (OpenAPI 3.2, com o método QUERY) é atualizado por `crud add`/`crud rm`.
  A API serve o spec em `/v1/openapi.yaml` e a UI (Scalar) em `/docs` pelo container `server`
  (contrato `contracts.IServerRepository`); `crud rm server` tira o container, e o spec continua sendo atualizado.

Nomes: `crud add` aceita singular ou plural (`component_definition`, `component_definitions`, `component-definitions`):

| o quê | exemplo |
|---|---|
| pasta (sem `_`, exceto `http_controllers`) | `internal/containers/componentdefinitions/{entities,repositories,usecases,http_controllers}` |
| packages (plural, com `_`) | `component_definitions`, `component_definitions_entities`, `_repositories`, `_usecases`, `_http_controllers` |
| tipos (sem `_`) | `ComponentDefinition`, `ComponentDefinitionsRepository`, `CreateComponentDefinitionUseCase` |
| rota / hacks | `/v1/component-definitions`, `hacks/http/component-definitions/` |
| tabela | `component_definitions` |

`devx crud add user` gera `internal/containers/users` (controller → usecase → repository), a migration,
`hacks/http/users/*.http` e o openapi:

| método | rota               | usecase        |
|--------|--------------------|----------------|
| POST   | /v1/users          | create         |
| GET    | /v1/users          | list           |
| QUERY  | /v1/users          | query (filtros, paginação e ordenação no body) |
| GET    | /v1/users/:id      | get_by_id      |
| PATCH  | /v1/users/:id      | update_by_id (parcial) |
| PUT    | /v1/users/:id      | upgrade_by_id (substituição total) |
| DELETE | /v1/users/:id      | delete_by_id (soft delete) |

## Como o devx edita o projeto
O código gerado não tem comentários nem marcadores. O devx encontra onde escrever pela estrutura:
- Go (`cmd/api/main.go`, `cmd/cli/main.go`, `pkg/dotenv/dotenv.go`, `internal/commands/root.go`): pela AST,
  ex: conexões antes do `workers.New(`/primeiro `X.Init(`, containers depois do último `X.Init(ctx, api`,
  campos dentro de `type Cfg struct`, validação antes do `return dotenvNullValue`.
- YAML (`infra/compose/docker-compose.yaml`, `docs/openapi/openapi.yaml`): pelo caminho das chaves
  (`services`, `services.api.environment`, `paths`, `components.schemas`).
- Makefile: os targets do devx (`run-dev`, `run-build`, `docker-restart`, `run-docker`, `view-*`, `sudo-clean`,
  `run-cli`, `build-cli`) são regerados a partir do estado do projeto; targets escritos à mão não são tocados.

Se você renomear algo que o devx usa como âncora (ex: `X.Init(ctx, api`, `api.Run(`, `defer cancel()`, `type Cfg struct`),
o comando falha avisando o que não encontrou.
