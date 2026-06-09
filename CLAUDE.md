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
- Toda nova função recebe um teste. Correções de bug recebem teste de regressão.
- Simule I/O externo (API, DB, filesystem) com interfaces nomeadas e fakes, não stubs inline.
- Testes devem ser F.I.R.S.T: rápidos, independentes, repetíveis, auto-validáveis, pontuais.

## Dependencies

- Injete dependências via construtor ou parâmetro, nunca via global ou init().
- Envolva libs de terceiros atrás de uma interface fina pertencente a este projeto.

## Structure

- Siga as convenções de pacotes Go: `cmd/`, `internal/`, `pkg/` quando aplicável.
- Prefira pacotes pequenos e focados a arquivos deus.
- Caminhos previsíveis: separe lógica de domínio, persistência e entrega em pacotes distintos.

## Formatting

- Use `gofmt` / `goimports`. Não discuta estilo além disso.

## Logging

- JSON estruturado para logs de debug / observabilidade.
- Texto simples apenas para saída CLI voltada ao usuário.
