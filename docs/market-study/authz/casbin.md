# Casbin (Apache Casbin)

| Campo | Valor |
|---|---|
| Categoria | Biblioteca de controle de acesso embutível, baseada em modelo configurável (PERM: request, policy, effect, matchers) |
| Site / repositório | https://casbin.org — https://github.com/apache/casbin (antigo `casbin/casbin`) |
| Licença e modelo de negócio | Apache 2.0. **Podling da Apache Incubator desde 07/02/2026** (voto de doação em #1625, abr/2026); primeira release oficial ASF "3.11.0-incubating" em votação (relatório jul/2026). Ecossistema da mesma comunidade inclui produtos como Casdoor (IAM) |
| Stack | Go (núcleo), *ports* "production-ready" em Java, Node.js, PHP, Python, .NET, C++, Rust; expressões via `casbin/govaluate`; adaptadores de armazenamento e *watchers* em pacotes separados |
| Maturidade | Desde 2017; ~20,5k estrelas, ~1,8k forks; Go v3.11.0 (ago/2026; módulo agora `github.com/apache/casbin/v3`); v2 ainda publicada (v2.135.0, dez/2025). Relatório da Incubator aponta 249 repositórios na org, "highly unusual" |
| Acesso | 2026-11-03 |

## Resumo (até 5 linhas)
Biblioteca que avalia `Enforce(sub, obj, act, ...)` contra uma lista de políticas (linhas CSV/DB) usando um *matcher* escrito em expressão. O modelo é configurável num arquivo `.conf` (ACL, RBAC, RBAC com domínios/tenants, ABAC, RESTful, deny-override, prioridade). Muito popular, leve (3 dependências diretas) e com dezenas de adaptadores de banco e watchers. Fraquezas: avaliação essencialmente linear sobre as políticas, sem avaliação parcial real, sem framework de testes/análise de políticas e histórico de problemas de concorrência.

## Funcionalidades por domínio (5.2)
| Domínio | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | ACL, RBAC (com domínios, papéis de recurso, *conditional role manager*), ABAC (campos de structs/JSON no matcher), RESTful, prioridade, deny-override. ReBAC só por papéis/grupos (`g`, `g2`) |
| 2. Decisões | ◐ | `Enforce`, `BatchEnforce`, `EnforceEx` (regra que casou); listagem por enumeração de políticas (`GetImplicitPermissionsForUser`, `GetImplicitUsersForResource`, `GetAllowedObjectConditions`); `Explain()` via LLM (v3.10) |
| 3. Políticas | ◐ | Modelo `.conf` + políticas CSV/DB; editor online; sem versionamento, testes declarativos ou simulação nativos |
| 4. Dados | ◐ | Adaptadores (dezenas de bancos/ORMs), *filtered policy loading*, watchers para sincronizar instâncias; sem tokens de consistência; issue aberta sobre transação junto com o banco da aplicação (#1702) |
| 5. Acesso dinâmico | ○ | Nada de primeira classe; *conditional role manager* permite papéis com condição (ex.: janela de tempo) |
| 6. Estados e transições | ◐ | Estado como atributo no matcher ABAC; sem modelo de transições |
| 7. Padrões | ○ | Sem AuthZEN (ausente do interop), sem RAR/UMA; issue de XACML fechada (#64) |
| 8. Máquinas e agentes | ○ | Nada específico no núcleo (projetos satélites da org para agentes, ex.: "Casbin Gateway") |
| 9. Desempenho e resiliência | ◐ | Embutido e local; desempenho degrada com milhares de regras (ver abaixo); `CachedEnforcer`, `SyncedEnforcer` |
| 10. Auditoria | ◐ | Interface de logger (reformulada em v3.6–v3.7); sem decision log estruturado |
| 11. Experiência e administração | ◐ | Editor online; UIs de administração vêm de projetos satélites (Casdoor etc.) |

## Diferenciais / exclusivos
- **Modelo de acesso como configuração** (PERM): troca-se de ACL para RBAC com domínios mudando o `.conf`, sem mexer no código.
- **Multi-linguagem**: mesmo modelo/política executado em 8 linguagens.
- **Ecossistema de adaptadores e watchers**: Postgres, Redis, etcd, TiKV, Kafka, NATS, ZooKeeper, Go Cloud (Pub/Sub, SNS/SQS, RabbitMQ) etc.
- `GetAllowedObjectConditions`: devolve condições de objeto armazenadas na política (ex.: `price < 25`) para filtragem em ORM — atalho simples, não avaliação parcial.
- Detecção de ciclos em hierarquia de papéis (DefaultDetector, DFS, v3.5.1).

## O que a comunidade elogia (com evidência)
- Embutível na aplicação Go, sem sistema externo; plugins Postgres/Redis facilitam (HN #26934520, 2021).
- Fácil de evoluir a camada de permissões sem mudar código; ABAC em Node.js implementado "em um dia" (HN #26934520).
- Popularidade e *ports* (estrelas; README).

## Do que a comunidade se queixa (com evidência)
- **Desempenho** com matchers de padrão: "alguns milhares de regras arruinaram nosso tempo de resposta" (pycasbin + keyMatch, HN #26934520). Benchmark oficial confirma crescimento linear (abaixo).
- **Complexidade/docs**: "parece complicado demais", "docs não eram boas, tive de usar o debugger" (HN #26934520).
- **Confiança**: logos de "usuários" no site sem comprovação (HN #26934520).
- **Concorrência/corretude**: data race em `RoleManagerImpl.HasLink()` (#1318, 2025), escritas concorrentes em assertion (#1514, 2025), `DistributedEnforcer` sem locks (#1022).
- **Listagem**: pedido de avaliação parcial como no OPA (#545) fechado com remissão a outra issue, sem implementação; govaluate não suporta resultados parciais.
- Escala de políticas com milhões de recursos (#700).

## Histórico de segurança
| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| (nenhum GHSA/CVE da biblioteca) | — | Página de advisories vazia; OSV sem entradas para `casbin/casbin` (v1/v2) | Ausência de advisories não prova ausência de falhas: bugs de corretude/concorrência foram tratados como issues comuns |
| Correção de quoting CSV em `SavePolicy` (v3.11.0) | 2026 | Campos com vírgula corrompiam/reinterpretavam a política ao salvar | Formatos de política textuais exigem escape canônico e testes de *round-trip* |
| CVE-2024-41657, CVE-2025-61524 (Casdoor, projeto irmão) | 2024–25 | Comparação incorreta; bypass de autorização por concatenação de URL | Casar caminhos/URLs por padrão (keyMatch) é fonte clássica de bypass; normalizar antes de autorizar |

## Arquitetura e decisões notáveis
- Políticas são tuplas genéricas (`p, sub, obj, act`), o *matcher* é uma expressão avaliada para cada linha candidata — simplicidade em troca de desempenho.
- Papéis em grafo em memória (role manager) com funções de casamento (keyMatch, regexMatch, globMatch) e domínios.
- Escrita/carregamento via adaptadores; *auto-save*; watchers notificam outras instâncias para recarregar (consistência eventual).
- `EnableAcceptJsonRequest`: JSON como atributo com custo 1,1–1,5x.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: "RBAC with domains/tenants" (`g = _, _, _`), um enforcer por tenant ou domínio como dimensão.
- Extensibilidade: funções customizadas no matcher, role managers, adaptadores, watchers, dispatchers.
- Bancos: praticamente todos via adaptadores (SQL via GORM/XORM/ent, Mongo, Redis, DynamoDB, etc.).
- Soberania: biblioteca local; dados onde o usuário quiser.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Embutibilidade em Go:** o ponto mais forte. v3 com **3 dependências diretas** (doublestar, govaluate, uuid) [proxy.golang.org]. Atenção à mudança de caminho do módulo para `github.com/apache/casbin/v3`.
- **Linguagem:** arquivo `.conf` com seções + expressões govaluate; políticas em CSV/DB. Não há linguagem de políticas de alto nível nem schema tipado.
- **Consistência:** eventual entre instâncias (watchers); nenhuma garantia transacional com o banco da aplicação (#1702).
- **Listagem:** sem query plan; APIs de enumeração sobre políticas e papéis (`GetImplicit*`), `GetAllowedObjectConditions` (condições literais por prefixo `r.obj.`), *filtered policy* para carregar subconjuntos.
- **Explicabilidade:** `EnforceEx` devolve a regra que casou; `Explain()` por LLM (v3.10) — explicação não determinística, cuidado para auditoria.
- **Testes:** não há framework de testes de políticas; testa-se com testes unitários da linguagem hospedeira; editor online para experimentar.
- **Análise estática:** detecção de ciclos de papéis (v3.5); nenhuma verificação de tipos ou de equivalência.
- **Desempenho publicado (Go, i7-6700HQ):** ACL 2 regras 0,015 ms/op; RBAC 1.100 regras 0,16 ms; **11.000 regras 2,26 ms; 110.000 regras 23,9 ms e 7,6 MB por chamada** [casbin.org/docs/benchmark — fonte do projeto, hardware antigo]. Python 3–4x mais lento.
- **AuthZEN:** não encontrado (interop AuthZEN, docs, busca web).

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N)? **Não.** Não há conceito de aprovação; nem a contagem é idiomática no matcher.
- Autorização sensível ao estado / transições declarativas? **Parcial (fraco).** Possível comparar atributo de estado no matcher ABAC; sem modelo de estados/transições.

## Implicações para o projeto
1. **Adotar** a ideia de "modelo como configuração" (premissa 3/18): o cliente escolhe RBAC/ABAC/domínios por configuração — mas com validação e schema, que faltam à Casbin.
2. **Evitar** avaliação por varredura linear de regras: a curva 0,16 ms → 23,9 ms (1k → 110k regras) mostra o limite; indexar por principal/recurso/ação desde o início.
3. **Evitar** explicação de decisões por LLM como fonte de verdade; explicabilidade deve ser determinística e auditável.
4. **Adotar** detecção de ciclos em hierarquias e testes de concorrência (`-race`, fuzz) — os bugs mais votados da Casbin são de concorrência.
5. **Oportunidade:** listagem eficiente e testes de políticas são lacunas antigas e votadas (#545, #700); bibliotecas Go leves com essas capacidades praticamente não existem.
6. **Adotar** watchers plugáveis (Postgres LISTEN/NOTIFY, NATS etc.) para invalidar caches entre instâncias, com interface única (premissa 9).

## Fontes
- https://github.com/apache/casbin — 2026-11-03
- https://github.com/apache/casbin/releases (páginas 1 e 2) — 2026-11-03
- https://github.com/apache/casbin/security/advisories — 2026-11-03
- https://osv.dev (Go `github.com/casbin/casbin`, `/v2`) — 2026-11-03
- https://github.com/apache/casbin/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://github.com/apache/casbin/issues/545 — 2026-11-03
- https://incubator.apache.org/clutch/casbin.html — 2026-11-03 (visto via busca)
- https://cwiki.apache.org/confluence/spaces/INCUBATOR/pages/435028221/July2026 — 2026-11-03 (visto via busca)
- https://proxy.golang.org (`github.com/casbin/casbin/v3`, `/v2`, `github.com/apache/casbin/v3`) — 2026-11-03
- https://casbin.org/docs/benchmark — 2026-11-03 — fonte do projeto
- https://casbin.org/docs/rbac-api e /docs/management-api — 2026-11-03
- https://casbin.org/docs/abac e /docs/watchers — 2026-11-03
- https://news.ycombinator.com/item?id=26934520 — 2026-11-03 — comunidade
- https://cvefeed.io/vuln/detail/CVE-2025-61524 — 2026-11-03 — Casdoor, não a biblioteca (visto via busca)
- https://authzen-interop.net/docs/intro (lista de implementações; Casbin ausente) — 2026-11-03
