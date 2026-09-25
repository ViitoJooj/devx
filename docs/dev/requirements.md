# Requisitos funcionais.

[x] `devx init <project>` gera a estrutura base (noxacloud).
[x] `devx crud add <name>` gera container com create, list, query, get, update(PATCH), upgrade(PUT) e delete.
[x] `devx service add <service>` para postgres, redis, rabbitmq e resend.
[x] `devx service add` para stripe, prometheus e grafana.
[x] `devx worker add <name> [--interval 1h]` gerando internal/workers/<name>.
[x] OpenAPI gerado e atualizado pelo `crud` (docs em /docs).
[x] `devx service add docker` (Dockerfile + api no compose).
[x] `rm` e `ls [--json]` para service, crud e worker (padrão gh: substantivo + verbo).
[x] `devx view add|rm|ls` para tauri, tauri-svelte, tauri-react, sveltekit, react e next (www/<nome>, CORS, makefile).
[x] `devx init --api|--cli` (monorepo: cmd/api + cmd/cli), `devx app add|rm|ls` (rm com --sudo), viper como service da cli e `devx command add|rm|ls`.
[x] Crud sem banco: pergunta postgres/memória, `--db`, `devx crud db <name> postgres|memory`.
[x] Postgres com privilégio mínimo: usuário da API só com SELECT/INSERT/UPDATE, migrations com o dono (sem pasta scripts/).
[x] `devx crud grant|revoke <name> <operação...>` e coluna ACCESS no `crud ls` (permissões por tabela via migration).
[ ] Cliente tipado do OpenAPI nas views (openapi-typescript + openapi-fetch).

# Adiado.

[ ] Campos na entidade via flag (ex: `devx crud add user --name string min=3 max=255 unique not-null --email string`).
[ ] `devx migrate compose-kubernetes` e `devx migrate kubernetes-compose` (YAML puro + kustomize em infra/k8s).
[ ] `devx migrate postgres-mongo`, `postgres-oracle` e inversos (trocar o banco principal do projeto).
