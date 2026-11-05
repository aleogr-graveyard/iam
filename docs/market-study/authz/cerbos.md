# Cerbos

| Campo | Valor |
|---|---|
| Categoria | PDP de políticas sem estado (RBAC/ABAC/PBAC) como serviço/sidecar + plano de controle SaaS (Cerbos Hub) |
| Site / repositório | https://www.cerbos.dev — https://github.com/cerbos/cerbos — docs: https://docs.cerbos.dev |
| Licença e modelo de negócio | PDP Apache 2.0, "free forever". **Cerbos Hub** pago por *monthly active principals*: PoC US$0 (até 100 MAP), Development a partir de US$25/mês, Production a partir de US$933/mês (5.000 MAP), Enterprise (SSO, Hub *self-hosted*, retenção de auditoria, SLA). **Cerbos Synapse** (enriquecimento de contexto) é produto novo do lado pago. Empresa: Zenauth Ltd (Reino Unido) |
| Stack | Go; condições em CEL (`cel-go`); gRPC + HTTP (protobuf); políticas YAML/JSON; Badger, SQLite, Postgres, MySQL como *stores* opcionais |
| Maturidade | Desde 2021; ~4,6k estrelas; v0.57.0 (23/10/2026), ainda sem 1.0; releases a cada 1–2 meses |
| Acesso | 2026-11-03 |

## Resumo (até 5 linhas)
PDP sem estado focado em autorização de aplicações: políticas de recurso, de principal, de papel e *derived roles* em YAML com condições CEL; a aplicação envia principal e recursos com atributos. Destaques: **PlanResources** (query planner que devolve AST de condição para filtrar listagens, com adaptadores de ORM), **suporte nativo a AuthZEN 1.0** (v0.48), suíte de testes declarativa e políticas com escopos hierárquicos para multi-tenant. Hub adiciona CI de políticas, playground, distribuição, auditoria agregada e PDP embutido em Wasm.

## Funcionalidades por domínio (5.2)
| Domínio | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC, ABAC, PBAC, derived roles; ReBAC apenas via atributos enviados (sem grafo) — Synapse busca relações em sistemas externos (pago) |
| 2. Decisões | ● | `CheckResources` (vários recursos × várias ações por chamada), `PlanResources` (listagem), outputs; explicação por trace/`cerbos repl` e audit log com políticas efetivas |
| 3. Políticas | ● | YAML/JSON, versões de política (`version`), escopos, `cerbos compile` com testes, schemas JSON para atributos; implantação gradual não nativa |
| 4. Dados | ◐ | Sem armazenamento de dados de negócio; atributos por requisição; Synapse (pago) para enriquecer |
| 5. Acesso dinâmico | ○ | Nada de primeira classe; JIT via condição de tempo (`now()`) |
| 6. Estados e transições | ◐ | Estado como atributo do recurso em condições CEL; sem modelo de transições |
| 7. Padrões | ● | AuthZEN Authorization API 1.0 nativo (evaluation, evaluations, metadata); participa do interop inclusive do perfil de busca |
| 8. Máquinas e agentes | ◐ | Autorização de serviços; marketing para MCP/agentes de IA; nada específico no PDP |
| 9. Desempenho e resiliência | ● | Sidecar/DaemonSet/Lambda, sem estado; ePDP Wasm (Hub) para edge/cliente |
| 10. Auditoria | ● | Access logs e decision logs com filtros; backends local, arquivo, Kafka, Hub; `cerbosCallId` por resposta |
| 11. Experiência e administração | ◐ | Playground, IDE colaborativo e gestão no Hub (pago); SDKs em 8 linguagens; sem UI no OSS |

## Diferenciais / exclusivos
- **Query planner (`PlanResources`)**: resposta `KIND_ALWAYS_ALLOWED`, `KIND_ALWAYS_DENIED` ou `KIND_CONDITIONAL` + AST da condição; adaptadores de referência para Prisma, Drizzle, Mongoose, Convex, LangChain/ChromaDB e SQLAlchemy [docs API].
- **Escopos hierárquicos** (`scope: acme.hr.uk`) com `scopePermissions` para sobrescrever políticas por tenant/organização — herança de políticas por tenant de primeira classe.
- **Testes declarativos** em YAML com *fixtures* de principais, recursos e auxData, relógio fixo (`now`), matriz de expectativas e outputs; executados por `cerbos compile` (inclusive no CI do Hub).
- AuthZEN nativo, inclusive participação no cenário de **busca** do interop (198/198 casos aprovados).
- ePDP: políticas compiladas para Wasm e distribuídas por CDN pelo Hub, com regras de quais políticas vão para quais clientes.

## O que a comunidade elogia (com evidência)
- Separação clara de políticas em YAML legível e GitOps (README; "Show HN" 2024 — pouca discussão: 24 pontos, 3 comentários).
- Integração com ORMs para listagem (adaptadores de query plan) — feature mais citada em comparativos (fonte: própria Cerbos — viés).
- Não encontrado: volume relevante de discussões independentes (HN, Reddit) — procurado em HN Algolia e busca web.

## Do que a comunidade se queixa (com evidência)
- Não embutível como biblioteca Go no OSS; a resposta oficial para "embutido" é ePDP Wasm via Hub (pergunta no HN #42307742; página ePDP).
- Issues mais votadas: cobertura de testes (#1654, aberta desde 2023), auditoria para NATS (#1722), erro interno do PlanResources com lista vazia (#2820), Git driver sem Azure DevOps (#1279), SASL para Kafka (#2348).
- **Correções de query plan permissivo demais** em releases recentes: v0.57.0 "over-permissive query plans for resource policies with only DENY rules for a role"; v0.56.0 precedência de principal policy no planner; v0.53.0 planner ignorava `OVERRIDE_PARENT` para DENY de escopo pai — sem advisory publicado.
- Preço por MAP no Hub pode crescer rápido (US$933/mês no Production) — observação nossa a partir da página de preços.

## Histórico de segurança
| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| (nenhum GHSA/CVE do projeto) | — | Página de advisories vazia; OSV sem entradas para `github.com/cerbos/cerbos`; só advisories de distribuições (Chainguard) por stdlib Go desatualizada | Atualização contínua de toolchain é parte da segurança |
| Bugs de planner (v0.53, v0.56, v0.57) | 2026 | Filtro de listagem mais permissivo que o `Check` (divergência) | **Testes diferenciais obrigatórios entre `Check` e listagem**; tratar divergência como vulnerabilidade com advisory |
| Remoção do cache de verificação JWT (v0.53) e endurecimento de JWT | 2026 | Validação de tokens no PDP | Validar identidade fora do motor ou com o mesmo rigor do AuthN |

## Arquitetura e decisões notáveis
- PDP sem estado; políticas compiladas para uma "rule table" em memória (roaring bitmaps, constant folding no CEL desde v0.52–v0.55).
- Fontes de política: disk, git (preferido), blob (S3/GCS/Azure/Minio), SQLite/Postgres/MySQL, Hub; recarga ao detectar mudanças.
- Modo estrito de avaliação (v0.55) e rejeição de `scopePermissions` conflitantes (v0.56): endurecimento progressivo.
- Telemetria anônima ligada por padrão (desligável).
- go.mod do servidor com ~275 requisitos (AWS SDK, go-git, Badger, CEL, migrate…) [proxy.golang.org v0.57.0].

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: escopos hierárquicos por tenant; Hub com workspaces e ePDP Rules por cliente/tenant.
- Extensibilidade: funções CEL próprias da Cerbos; sem plugins de usuário.
- Configuração: YAML; bancos só para armazenar políticas (SQLite, Postgres, MySQL).
- Soberania: PDP roda onde o usuário quiser; Hub é SaaS (Enterprise oferece Hub *self-hosted*).

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Embutibilidade em Go:** não oficial. O SDK Go (`cerbos-sdk-go`) é cliente gRPC; importar o motor exige o módulo do servidor (dependências pesadas, API sem garantia, versão 0.x).
- **Linguagem:** YAML + CEL; tipos de política: resource, principal, role, derived roles, exported variables/constants.
- **Consistência:** eventual na distribuição de políticas; dados vêm na requisição (sem *new enemy problem* interno, ele é transferido à aplicação).
- **Listagem:** `PlanResources` (AST → adaptador de ORM); simplificação de `all()` em AND (v0.48). Não lista "quais usuários acessam X".
- **Explicabilidade:** decision logs com políticas e derived roles efetivas; trace detalhado (formato TraceBatch, v0.52).
- **Testes:** suites YAML, filtro de testes (v0.51), `now` fixo, outputs; cobertura ainda não (#1654).
- **Análise estática:** compilação com verificação de schemas e referências; *permissions advisor* (v0.52); sem verificação formal.
- **Desempenho publicado:** apenas qualitativo — "sub-millisecond", "thousands of checks per second" [docs Hub]. Números de terceiros não confirmados (página Mintlify: P50 < 5 ms — não oficial). Não encontrado benchmark reprodutível oficial.
- **AuthZEN:** nativo desde v0.48 (09/12/2025), descrito na doc como "partially implements"; mapeia para `CheckResources`.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N)? **Não.** Possível apenas como condição CEL sobre atributos enviados (ex.: número de aprovações), sem fluxo, prova ou vinculação à ação.
- Autorização sensível ao estado / transições declarativas? **Parcial.** Condições sobre `R.attr.status` são o padrão dos exemplos; ações podem representar transições, mas não há declaração de estados/transições.

## Implicações para o projeto
1. **Adotar** escopos hierárquicos com regras explícitas de sobrescrita para políticas por tenant/organização (5.2 item 3 e 4.7).
2. **Adotar** o formato de resposta de listagem em três estados (sempre permitido / sempre negado / condicional + AST) e adaptadores de consulta como pacotes separados.
3. **Adotar** testes declarativos de políticas com *fixtures* e relógio fixo; acrescentar cobertura (lacuna da Cerbos).
4. **Aprender** com os bugs de planner: listagem e checagem precisam de testes diferenciais/propriedade contínuos; divergência vira advisory.
5. **Adotar** AuthZEN nativo cedo, incluindo os perfis de busca — a Cerbos mostra que é viável e vira argumento de venda.
6. **Oportunidade:** Cerbos não é embutível em Go no OSS e cobra o PAP; um produto com motor embutível + UI administrativa aberta preenche esse espaço.
7. **Evitar** precificação apenas por principal ativo sem previsibilidade (queixa potencial; validar no estudo de preços).

## Fontes
- https://github.com/cerbos/cerbos — 2026-11-03
- https://github.com/cerbos/cerbos/releases (v0.51–v0.57) — 2026-11-03
- https://github.com/cerbos/cerbos/security/advisories — 2026-11-03
- https://osv.dev (Go `github.com/cerbos/cerbos`) e resultados de busca CGA (Chainguard) — 2026-11-03
- https://github.com/cerbos/cerbos/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://www.cerbos.dev/pricing — 2026-11-03 — fonte do fornecedor
- https://www.cerbos.dev/blog/cerbos-pdp-v0-48-open-id-auth-zen-support-improved-query-plans-faster-bundle-loading — 2026-11-03 — fornecedor
- https://www.cerbos.dev/features-benefits-and-use-cases/wasm-embedded-pdp.md — 2026-11-03 — fornecedor
- https://docs.cerbos.dev/cerbos/latest/api/index.html (CheckResources, PlanResources, AuthZEN) — 2026-11-03
- https://docs.cerbos.dev/cerbos/latest/policies/compile.html (testes) — 2026-11-03
- https://docs.cerbos.dev/cerbos/latest/policies/scoped_policies.html — 2026-11-03
- https://docs.cerbos.dev/cerbos/latest/configuration/storage.html e /audit.html — 2026-11-03
- https://docs.cerbos.dev/cerbos-hub/decision-points — 2026-11-03
- https://proxy.golang.org/github.com/cerbos/cerbos/@v/v0.57.0.mod — 2026-11-03
- https://authzen-interop.net/docs/scenarios/search/results/cerbos/ e /todo-1.1/results/cerbos/ — 2026-11-03
- https://news.ycombinator.com/item?id=42307742 — 2026-11-03 — comunidade
- https://www.mintlify.com/cerbos/cerbos/operations/performance — 2026-11-03 — não oficial, não confirmado (visto via busca)
