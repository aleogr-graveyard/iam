# OPA (Open Policy Agent) / Rego

| Campo | Valor |
|---|---|
| Categoria | Motor de políticas de uso geral (PBAC/ABAC), embutível como biblioteca Go ou executado como servidor/sidecar |
| Site / repositório | https://www.openpolicyagent.org — https://github.com/open-policy-agent/opa |
| Licença e modelo de negócio | Apache 2.0, projeto **CNCF Graduated**. Sem empresa comercial por trás desde ago/2025: o time da Styra (criadores) foi para a Apple; produtos comerciais da Styra (DAS) descontinuados ou repassados à comunidade (ver "Situação pós-Styra") |
| Stack | Go. Políticas em Rego (Datalog estendido). Dados em memória (opcional em disco com Badger). Também compila para Wasm. Swift OPA (avaliador nativo em Swift) |
| Maturidade | Desde 2016; ~12,4k estrelas, ~1,7k forks; release mensal (v1.22.0 em 29/10/2026). OPA 1.0 em dez/2024. Usos públicos: Apple, Reddit, Netflix, Kubernetes (Gatekeeper) |
| Acesso | 2026-11-03 |

## Resumo (até 5 linhas)
Motor genérico de "policy as code": recebe `input` JSON + `data` JSON, avalia Rego e devolve qualquer JSON (não só allow/deny). Domina infraestrutura (Kubernetes, Terraform, Envoy) e é usado também em autorização de aplicações. Embutível em Go (`rego`, `sdk`), com testes, cobertura, benchmark, linter (Regal) e avaliação parcial que gera filtros SQL/UCAST. Pontos fracos para nosso caso: curva de aprendizado de Rego, ausência de modelo ReBAC nativo, sincronização de dados por bundles (consistência eventual) e dependências pesadas.

## Situação pós-Styra (ago/2025 → nov/2026)
- 20/08/2025: os criadores (Teemu Koponen, Tim Hinrichs, Torin Sandall) e "muitos membros da Styra" entraram na Apple. Governança e licença inalteradas; lista de mantenedores inalterada, só mudou a empresa [OPA blog].
- Repositórios da Styra encaminhados à org CNCF/OPA: **EOPA** (distribuição comercial com *data filtering*), **OPA Control Plane** (constrói bundles a partir de git e publica em S3/GCS/Azure), SDKs (TS, React, UCAST‑Prisma, C#, ASP.NET, Java, Spring Boot) e **Regal** (linter) [OPA blog].
- Rego Playground continua operado pela Styra [OPA blog]. Styra DAS: tratado como descontinuado por concorrente [Oso — viés: concorrente].
- Evidência de continuidade: releases mensais seguiram (v1.19.1 a v1.22.0 entre ago e out/2026), correção de segurança em out/2026, pesquisa da comunidade 2026 aberta [GitHub releases; site OPA].
- Risco: sem fornecedor comercial de suporte; mantenedores centrais empregados por um único usuário final (Apple). Precedente citado: FoundationDB [Oso — viés: concorrente].
- A funcionalidade de *data filtering* (antes só no EOPA) aparece hoje na documentação oficial do OPA (Compile API com saída SQL/UCAST) [docs OPA].

## Funcionalidades por domínio (5.2)
| Domínio | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC/ABAC/PBAC por código Rego; ReBAC só com `graph.reachable` sobre dados carregados — sem modelo relacional nem índice de grafo |
| 2. Decisões | ◐ | Check individual; sem API REST de lote (issue #6625 aberta); listagem via Compile API (avaliação parcial → SQL/UCAST); explicação por *trace* (`explain=full`) e `--var-values` |
| 3. Políticas | ● | Rego, bundles versionados/assinados, `opa test`, cobertura, `opa bench`, `opa check --strict`, type checking com JSON Schema, Regal. Sem implantação gradual nativa |
| 4. Dados | ◐ | `input` por requisição + `data` via bundles/push API/`http.send`; sem tokens de consistência |
| 5. Acesso dinâmico | ○ | Nada de primeira classe (JIT, solicitações, M de N, delegação); obrigações expressáveis pelo retorno JSON livre |
| 6. Estados e transições | ◐ | Expressável por atributos (`input.resource.status`); sem declaração de máquina de estados |
| 7. Padrões | ◐ | AuthZEN apenas via plugin de terceiros/proxy; sem RAR/UMA nativos |
| 8. Máquinas e agentes | ◐ | Envoy ext_authz, Kubernetes, uso crescente em agentes/MCP pela comunidade; nada específico |
| 9. Desempenho e resiliência | ● | Biblioteca Go, sidecar, Wasm; indexação de regras; avaliação local é o modelo padrão |
| 10. Auditoria | ● | Decision logs com mascaramento, envio remoto, back-pressure |
| 11. Experiência e administração | ◐ | Sem UI de administração no OSS (DAS era o PAP); Playground, plugin VS Code, OPA Control Plane |

## Diferenciais / exclusivos
- Linguagem genérica: a mesma ferramenta para Kubernetes, Terraform, CI, APIs e aplicações.
- **Compile API** com alvos `postgresql`, `mysql`, `sqlserver`, `sqlite` e UCAST (`all`, `minimal`, `prisma`, `linq`), via cabeçalho `Accept` ou multitarget; políticas de filtro anotadas com `# METADATA compile: unknowns: [...]` [docs OPA REST API / Filtering].
- Bundles assinados + decision logs + status API: plano de distribuição desacoplado do motor.
- Ecossistema: Gatekeeper, Conftest, Regal, plugins de IDE, Swift OPA.

## O que a comunidade elogia (com evidência)
- Testabilidade das políticas, API melhor que Casbin, "individualmente testáveis" (HN #26934520, 2021).
- 68% aprenderam Rego em menos de uma semana (pesquisa OPA 2020 — viés: projeto).
- Graduação CNCF e avaliação conjunta de segurança da CNCF TAG Security.

## Do que a comunidade se queixa (com evidência)
- **Curva de aprendizado de Rego** (estilo Datalog/Prolog): 16% dizem ter dificuldade (pesquisa 2020); a avaliação de segurança da CNCF registra "non-trivial learning curve"; "No one wants to write Rego" (Permit.io — viés: concorrente).
- Falta de documentação para autorização de aplicações/IAM: issue #2098 é a mais votada em aberto (desde 2020).
- Sincronização de dados com o OPA trabalhosa ("diff pushing" com ciclos de vida separados — HN #26934520).
- Desempenho de carga/compilação com muitas políticas: #7715 (operações REST lentas com muitas políticas), #8911 (custo O(n²) na ativação de múltiplos bundles, 2026), #2282.
- Falta de API de lote (#6625) e de avaliação parcial em Wasm (#3407).
- Incerteza de suporte comercial pós‑Styra (Oso — viés: concorrente).

## Histórico de segurança
| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVE-2022-23628 / GHSA-hcw3-j74m-qc58 | 2022 | Formatação de código gerado muda ordem de arrays (semântica alterada) | Ferramentas que reescrevem políticas (formatter, migração) precisam de testes de equivalência semântica |
| CVE-2022-28946 | 2022 | Acesso fora dos limites (DoS) no parser | *Fuzzing* contínuo de parser de políticas |
| CVE-2022-33082 | 2022 | DoS no compilador | Limites de recursos na compilação de políticas fornecidas por tenants |
| CVE-2022-36085 / GHSA-f524-rf33-2jjr | 2022 | `with` permitia *mock* de builtins proibidos (bypass de `WithUnsafeBuiltins`) | Recursos de teste não podem existir no caminho de produção |
| CVE-2024-8260 | 2024 | Windows: caminho UNC força autenticação SMB (vazamento de hash NTLM) | Validar caminhos de arquivos de políticas/bundles |
| CVE-2025-46569 / GHSA-6m8w-jc87-6cr7 | 2025 | Injeção de Rego pelo caminho HTTP da Data API | Nunca concatenar entrada do usuário em consultas de política |
| GHSA-g2pf-6v5v-qcc8 | 2026 | Bundles não assinados (inclusive delta) aceitos quando a verificação estava configurada sem `keyid` | Falha fechada na verificação de assinatura de políticas; testar configurações parciais |

Fonte: GitHub Security Advisories do repositório e OSV.dev.

## Arquitetura e decisões notáveis
- Desacoplamento total do modelo: OPA não conhece "usuário", "papel" ou "recurso"; tudo é JSON.
- Avaliação em memória; JSON carregado ocupa ~20x o tamanho serializado (8 MB → ~160 MB); 10k regras ACL ≈ 130 MB, 100k ≈ 1,1 GB [docs Policy Performance].
- Avaliação de uma consulta é single-thread; ao embutir como biblioteca, a concorrência é responsabilidade do chamador [docs].
- "Fragmento linear" + indexação de regras para tempo próximo de constante; indexação de compreensões.
- Partial evaluation (desde 2018) é a base tanto de otimização quanto de *data filtering*.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: por pacotes/bundles/instâncias; nada de primeira classe.
- Extensibilidade: builtins customizados em Go, plugins (ex.: AuthZEN de terceiros), `http.send`.
- Configuração: YAML/JSON; bundles de HTTP, OCI, S3 etc.
- Bancos: não há; dados em memória ou disco local (Badger). Soberania: tudo roda na infraestrutura do usuário.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Embutibilidade em Go:** `github.com/open-policy-agent/opa/v1/rego` e `.../sdk`. Custo: go.mod com ~103 requisitos (OpenTelemetry, Prometheus, Badger, wazero, jwx, cobra…) [proxy.golang.org, v1.22.0]. Para nós (premissa 24), dependência pesada.
- **Linguagem:** Rego v1 (`if`, `contains`, `every`); v1.20 adicionou `and`/`or`; v1.22 introduziu import `rego.v2`.
- **Consistência:** eventual (bundles com *polling* ou push); sem "zookie"/token de consistência.
- **Listagem ("query plan"):** Compile API (`POST /v1/compile/{path}`) com unknowns declarados; retorna condição residual em SQL ou UCAST; políticas precisam caber num subconjunto traduzível (erros de fragmento reportados, v1.22). Não resolve "quais usuários acessam X" de forma geral; é filtro por condição.
- **Explicabilidade:** traces (`explain=notes|full`), `--var-values` em testes; v1.21 permite retornar rótulos de regras nas APIs Data/Query.
- **Testes:** regras `test_*` em Rego, `with` para *mock* de `input`/`data`/funções, cobertura, benchmark com saída `gobench` para `benchstat`.
- **Análise estática:** type checker com JSON Schema, `opa check --strict`, Regal (lint). Sem verificação formal/SMT.
- **Desempenho publicado:** exemplo de `opa bench` RBAC ~45 µs/op; testes de unidade RBAC ~26 µs/op, ~12 KB e ~230 alocações/op [docs]. Comparativo do paper do Cedar: Rego 42,8x–80,8x mais lento que Cedar (viés: autores do Cedar/AWS).
- **AuthZEN:** sem suporte nativo. Plugin de terceiros `kanywst/opa-authzen-plugin` (listado no ecossistema OPA); no interop AuthZEN, implementação "OPA" via proxy passou 43/43 casos do cenário todo‑1.1.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N) como recurso de primeira classe? **Não.** Só é possível passar aprovações como `input` e escrever regra de contagem; não há coleta, vinculação à ação nem verificação de assinaturas.
- Autorização sensível ao estado / transições declarativas? **Parcial.** Estado como atributo do `input` é trivial; não existe declaração de estados/transições nem definição compartilhada com a aplicação.

## Implicações para o projeto
1. **Adotar** a ideia de Compile API: avaliação parcial que gera filtros para vários dialetos (SQL, UCAST) é o padrão de mercado para listagem em motores de políticas.
2. **Adotar** decision logs com mascaramento e bundles assinados; **aprender** com GHSA-g2pf: verificação de assinatura deve falhar fechada em qualquer configuração parcial.
3. **Evitar** linguagem de uso geral estilo Datalog para autores de políticas de aplicação: a queixa de curva de aprendizado é constante.
4. **Evitar** OPA como dependência embutida: ~100 requisitos no go.mod conflita com a premissa 24; se usado, isolar atrás de adaptador.
5. **Oportunidade:** risco de governança pós‑Styra (sem suporte comercial, PAP descontinuado) abre espaço para um PAP aberto com UI administrativa.
6. **Adotar** o modelo de testes (testes como políticas, *mock* de dados, cobertura, benchmark com formato padrão) — mas sem que o mecanismo de *mock* exista no caminho de produção (CVE-2022-36085).

## Fontes
- https://openpolicyagent.org/blog/note-from-teemu-tim-and-torin-to-the-open-policy-agent-community-2dbbfe494371 — 2026-11-03 — fonte oficial do projeto
- https://osohq.com/post/opa-maintainers-join-apple-oss-community-to-maintain-styra-products — 2026-11-03 — viés: concorrente (Oso)
- https://cloudnativenow.com/editorial-calendar/best-of-2025/apple-buys-styra-brains-opa-remains-open-2/ — 2026-11-03 — imprensa (visto apenas no resultado de busca)
- https://github.com/open-policy-agent/opa — 2026-11-03
- https://github.com/open-policy-agent/opa/releases — 2026-11-03
- https://github.com/open-policy-agent/opa/security/advisories — 2026-11-03
- https://osv.dev (consulta API por `github.com/open-policy-agent/opa`) — 2026-11-03
- https://www.openpolicyagent.org/docs/filtering e /docs/filtering/partial-evaluation — 2026-11-03
- https://www.openpolicyagent.org/docs/rest-api (Compile API) — 2026-11-03
- https://www.openpolicyagent.org/docs/policy-performance — 2026-11-03
- https://www.openpolicyagent.org/docs/policy-testing — 2026-11-03
- https://proxy.golang.org/github.com/open-policy-agent/opa/@v/v1.22.0.mod — 2026-11-03
- https://github.com/open-policy-agent/opa/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc — 2026-11-03
- https://news.ycombinator.com/item?id=26934520 — 2026-11-03 — comunidade
- https://openpolicyagent.org/blog/open-policy-agent-survey-summary-spring-2020-adaa46e61f0 — 2026-11-03 — viés: projeto (visto via busca)
- https://tag-security.cncf.io/community/assessments/projects/opa/ — 2026-11-03 (visto via busca)
- https://permit.io/blog/no-one-wants-to-write-rego — 2026-11-03 — viés: concorrente (visto via busca)
- https://pkg.go.dev/github.com/kanywst/opa-authzen-plugin — 2026-11-03 — terceiro (visto via busca)
- https://authzen-interop.net/docs/scenarios/todo-1.1/results/opa/ — 2026-11-03
- https://arxiv.org/abs/2403.04651 (comparativo Cedar × Rego × OpenFGA) — 2026-11-03 — viés: autores do Cedar (AWS)
