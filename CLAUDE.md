# my-backlog

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

### Fakes

- Fakes de `Repository` ficam em `internal/backlog/backlogtest.FakeRepository` — nunca recrie inline.
- Fakes de `MigrationStore` ficam em `migrator_test.go` (pacote `database`, não exportados).
- Para testar `database.Connect` sem Oracle: use `database.WithOpener(fakeOpener{})`.
- Fakes devem expor apenas o mínimo necessário para asserção: `Saved()`, `SaveCount()`, `FailWith()`.
  Não adicione métodos de conveniência que não são testados.

## Dependencies

- Injete dependências via construtor ou parâmetro, nunca via global ou init().
- Envolva libs de terceiros atrás de uma interface fina pertencente a este projeto.

## Structure

- Siga as convenções de pacotes Go: `cmd/`, `internal/`, `pkg/` quando aplicável.
- Prefira pacotes pequenos e focados a arquivos deus.
- Caminhos previsíveis: separe lógica de domínio, persistência e entrega em pacotes distintos.

## Architecture

O projeto tem três pacotes internos com responsabilidades estritamente separadas:

```
internal/backlog          — domínio: Backlog, Service, Repository, Validator, erros
internal/backlog/backlogtest — fakes de teste: FakeRepository (nunca importar em produção)
internal/database         — infraestrutura: conexão Oracle, migrations, schema_migrations
```

**Fluxo de dependência:** `database` não conhece `backlog`; `backlog` não conhece `database`.
O ponto de composição fica em `main()` (ainda não existe — virá em `cmd/`).

### Erros

Use os dois tipos canônicos do pacote `backlog`:

- `InputError{Kind, Field, Message}` — erro de entrada do usuário, não retentável.
  `Kind` mapeia para HTTP status: `KindInvalid` → 400, `KindNotFound` → 404.
- `InfraError{Op, Cause}` — falha de infraestrutura, retentável. Implementa `Unwrap()`.

Nunca retorne `errors.New` ou `fmt.Errorf` solto no domínio — use esses tipos para que
camadas superiores possam fazer `errors.As` e decidir código HTTP / retry.

### Migrations

- Arquivos SQL ficam em `internal/database/migrations/NNNN_nome.sql` e são embarcados no binário.
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

## Formatting

- Use `gofmt` / `goimports`. Não discuta estilo além disso.

## Logging

- JSON estruturado para logs de debug / observabilidade.
- Texto simples apenas para saída CLI voltada ao usuário.
