# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Rodar todos os testes (sem Oracle)
go test ./...

# Rodar um teste específico
go test ./cmd/server/ -run TestRequireAuth_ValidSid
go test ./internal/session/ -run TestStore_NewAndGet

# Testes de integração (requer Oracle configurado via .env)
go test ./... -tags integration

# Build do servidor
go build ./cmd/server/

# Rodar o servidor (lê .env automaticamente)
go run ./cmd/server/

# Formatar código
goimports -w .
```

O servidor escuta em `:8081` por padrão; sobrescreva com `ADDR=:3000`. Variáveis de conexão Oracle ficam em `.env` (veja `.env.example`): `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_SERVICE_NAME`, `DB_WALLET_PATH`.

## Code style

- Functions: 4–20 lines. Divida se for maior.
- Arquivos: menos de 500 linhas. Separe por responsabilidade.
- Uma coisa por função, uma responsabilidade por pacote (SRP).
- Nomes: específicos e únicos. Evite `data`, `handler`, `manager`.
  Prefira nomes que retornem menos de 5 hits no codebase.
- Tipos: explícitos. Sem `interface{}` / `any` desnecessário, sem funções não tipadas.
- Sem duplicação de código. Extraia lógica compartilhada em funções/pacotes.
- Early returns em vez de ifs aninhados. Máximo 2 níveis de indentação.
- Mensagens de erro devem incluir o valor ofensivo e o formato esperado.

## Comments

- Mantenha seus próprios comentários. Não os remova em refatorações — eles carregam intenção e origem.
- Escreva POR QUÊ, não O QUÊ. Evite `// incrementa contador` acima de `i++`.
- Godoc em funções/tipos exportados: intenção + um exemplo de uso.
- Referencie número de issue ou SHA de commit quando uma linha existe por causa de um bug específico ou restrição externa.

## Tests

- Testes rodam com: `go test ./...`
- Testes de integração (Oracle real) usam `//go:build integration` e são ignorados no `./...` padrão.
- Toda nova função recebe um teste. Correções de bug recebem teste de regressão.
- Simule I/O externo (API, DB, filesystem) com interfaces nomeadas e fakes, não stubs inline.
- Testes devem ser F.I.R.S.T: rápidos, independentes, repetíveis, auto-validáveis, pontuais.
- Testes que cruzam fronteiras de pacotes (ex: pacote A testa com tipos de B) devem usar `package A_test`, nunca `package A` — evita ciclos de importação.

### Fakes

- Fakes de `backlog.Repository` ficam em `internal/backlog/backlogtest.FakeRepository` — nunca recrie inline.
- Fakes de `user.Repository` ficam em `internal/user/usertest.FakeRepository` — nunca recrie inline.
- Fakes de `session.Sessions` ficam em `internal/session/sessiontest.FakeStore` — nunca recrie inline.
  Use `SeedSID(sid, userID)` para pré-semear sessões conhecidas; `New()` retorna IDs determinísticos (`"sid-1"`, `"sid-2"`, …).
- Fakes de `MigrationStore` ficam em `migrator_test.go` (pacote `database`, não exportados).
- Para testar `database.Connect` sem Oracle: use `database.WithOpener(fakeOpener{})`.
- Fakes devem expor apenas o mínimo necessário para asserção: `Saved()`, `SaveCount()`, `FailWith()`.
  Não adicione métodos de conveniência que não são testados.

## Dependencies

- Injete dependências via construtor ou parâmetro, nunca via global ou init().
- Envolva libs de terceiros atrás de uma interface fina pertencente a este projeto.
- Interfaces consumidas apenas por uma camada são definidas nessa camada (consumidora), não no pacote que as satisfaz.
  Exemplo: `SessionGetter` vive em `cmd/server`, não em `internal/session`.

## Structure

- Siga as convenções de pacotes Go: `cmd/`, `internal/`, `pkg/` quando aplicável.
- Prefira pacotes pequenos e focados a arquivos deus.
- Caminhos previsíveis: separe lógica de domínio, persistência e entrega em pacotes distintos.

## Architecture

Mapa completo de pacotes e responsabilidades:

```
internal/apperrors              — tipos canônicos de erro: InputError, InfraError, InputKind
internal/backlog                — domínio: Backlog, Service, Repository, Validator
                                  re-exporta tipos de apperrors como aliases (backlog.InputError etc.)
internal/backlog/backlogtest    — fakes de teste: FakeRepository (nunca importar em produção)
internal/database               — infraestrutura: conexão Oracle, migrations, UUID
internal/session                — sessões: interface Sessions, implementação MemoryStore
internal/session/sessiontest    — fakes de teste: FakeStore (nunca importar em produção)
internal/user                   — domínio: User, Service, Repository, Validator
internal/user/usertest          — fakes de teste: FakeRepository (nunca importar em produção)
cmd/server                      — entrega HTTP: rotas, handlers, middlewares, mapeamento de erros
```

**Fluxo de dependência permitido:**

```
cmd/server  →  internal/session
            →  internal/user
            →  internal/apperrors
internal/backlog   →  internal/apperrors
                   →  internal/database   (apenas oracle_repository — para NewUUID)
internal/user      →  internal/apperrors
                   →  internal/database   (apenas oracle_repository — para NewUUID)
internal/database  →  (sem dependências internas)
internal/session   →  (sem dependências internas)
internal/apperrors →  (sem dependências internas)
```

`internal/database` não importa nenhum pacote de domínio (`backlog`, `user`). Testes do pacote `database` que precisem cruzar essa fronteira usam `package database_test`.

### Erros

Os tipos canônicos vivem em `internal/apperrors` (não em `backlog`):

- `InputError{Kind, Field, Message}` — erro de entrada do usuário, não retentável.
- `InfraError{Op, Cause}` — falha de infraestrutura, retentável. Implementa `Unwrap()`.

Nunca retorne `errors.New` ou `fmt.Errorf` solto no domínio — use esses tipos para que
camadas superiores possam fazer `errors.As` e decidir código HTTP / retry.

**Mapeamento `InputKind` → HTTP status** fica exclusivamente em `cmd/server/errorwriter.go`:

```
KindInvalid      → 422 (UnprocessableEntity)
KindNotFound     → 404
KindConflict     → 409
KindUnauthorized → 401
InfraError       → 500 (logado no servidor; detalhes nunca chegam ao cliente)
```

Todo call site de erro HTTP usa `writeError(w, err)` — nunca duplique essa tabela.

### Session

- `session.Sessions` é a interface usada pelos handlers (`New`, `Get`, `Delete`).
- `session.MemoryStore` é a implementação em memória (criada com `session.NewMemoryStore()`).
- Handlers e middlewares recebem `session.Sessions`, nunca `*session.MemoryStore`.
- Middlewares que só precisam de `Get` recebem `SessionGetter` (definida em `cmd/server`), não `session.Sessions`.

### HTTP delivery (cmd/server)

- **Handlers são renderizadores puros**: não verificam autenticação inline.
  Autenticação é responsabilidade dos middlewares `RequireAuth` e `RedirectIfAuth`.
- **`RequireAuth`**: redireciona para `/login` se não há sessão válida.
- **`RedirectIfAuth`**: redireciona para `/home` se já há sessão válida.
- Interfaces consumidas só pela entrega (`SessionGetter`) são definidas em `cmd/server`, não em `internal/`.

### Migrations

- Arquivos SQL ficam em `internal/database/migrations/NNNN_nome.sql` e são embarcados no binário via `database.MigrationsFS`.
- Para adicionar uma migration: criar o arquivo SQL e rodar o teste `TestSchemaConstraintsMatchGoConstants`
  se a migration alterar colunas `VARCHAR2` cujos limites têm constante Go correspondente.
- `MigrationStore` é a interface de persistência do Migrator — use um fake nas unidades,
  `OracleMigrationStore` apenas em integração.
- `ErrDDLAppliedUntracked` indica estado irrecuperável (DDL commitou, registro não). Requer
  intervenção manual; nunca silencie esse erro.

### Schema drift

As constantes `MaxTitleLength` e `MaxDescriptionLength` em `internal/backlog/validator.go`
**devem sempre espelhar** os limites `VARCHAR2(N)` correspondentes no SQL. O teste
`TestSchemaConstraintsMatchGoConstants` guarda isso automaticamente — se falhar, atualize
os dois ao mesmo tempo.

### UUID

Geração de UUID sempre via `database.NewUUID()` — nunca reimplemente localmente.
Retorna UUID v4 com `crypto/rand`; erros de geração devem ser embrulhados em `InfraError{Op: "generate-id"}`.

## Formatting

- Use `gofmt` / `goimports`. Não discuta estilo além disso.

## Logging

- JSON estruturado para logs de debug / observabilidade.
- Texto simples apenas para saída CLI voltada ao usuário.
