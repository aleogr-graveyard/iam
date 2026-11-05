# SpiceDB / AuthZed

| Campo | Valor |
|---|---|
| Categoria | Motor relacional estilo Zanzibar ("banco de dados de permissões"), servidor; AuthZed é a empresa e oferta comercial |
| Site / repositório | https://authzed.com · https://github.com/authzed/spicedb · docs: https://github.com/authzed/docs |
| Licença e modelo de negócio | Apache-2.0 (open source desde 09/2021, sem mudanças de licença). Pago: AuthZed Cloud (a partir de US$ 2/h por permissions system), Dedicated Cloud (por vCPU reservada), Self-Hosted Enterprise (licença anual por região e vCPU). **Materialize, audit logging, FIPS, Datadog e Kafka API só nos planos pagos** |
| Stack | Go; CockroachDB, Cloud Spanner, PostgreSQL 15+, MySQL (não recomendado), memdb. gRPC + HTTP gateway; CEL nos caveats; WASM para o playground; operador Kubernetes |
| Maturidade | Desde 2021; **v1.57.0 em 20/10/2026** (proxy.golang.org). ~7,2 mil estrelas, ~440 forks, ~5,7 mil commits. Adotantes citados: IBM, Red Hat, GitPod, Netflix (parceiro de design dos caveats), GitHub (contribuiu o datastore MySQL) |
| Acesso | 2026-11-03 |

## Resumo
- Implementação Zanzibar mais fiel e mais madura: **ZedTokens** (zookies), consistência configurável por requisição, Watch API, dispatch distribuído com cache entre nós.
- **Caveats** (CEL) dão ABAC sobre ReBAC, com resultado tri-estado (`CONDITIONAL_PERMISSION` + lista do contexto que falta).
- Listagem (LookupResources/LookupSubjects) com cursor, mas o próprio fornecedor a declara "pesada" acima de ~10 mil resultados e empurra **Materialize**, pago e GA desde 09/2026.
- Embutível em Go: pacote `pkg/embedded` (só Check, sempre totalmente consistente) ou servidor em processo via bufconn.
- Sem AuthZEN; sem multi-tenancy nativa (issue #204 aberta).

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | ReBAC e RBAC no schema; ABAC via caveats CEL. Sem PBAC |
| 2. Decisões | ● | CheckPermission, CheckBulkPermissions, LookupResources, LookupSubjects, ReadRelationships, Expand. Explicação via tracing (`zed --explain`, debug trace) |
| 3. Políticas | ◐ | Linguagem de schema própria; `zed validate`, assertions e "expected relations", GitHub Actions; migrações online de schema. Um schema por permissions system (sem versões paralelas nem rollout gradual) |
| 4. Dados | ● | ZedTokens, 4 níveis de consistência, Watch com histórico desde um token, caveats com contexto parcial na escrita, expiração de relações, BulkImport/Export |
| 5. Acesso dinâmico | ◐ | **Relações com expiração** nativas e caveats temporais. Sem solicitações, aprovações nem M de N |
| 6. Estados e transições | ○ | Estado só como parâmetro de caveat |
| 7. Padrões | ○ | AuthZEN não encontrado (docs, issues, interop OpenID) |
| 8. Máquinas e agentes | ◐ | Sujeitos arbitrários; AuthZed se posiciona em "AI Authorization" (RAG, agentes). Nada de MCP no núcleo verificado |
| 9. Desempenho e resiliência | ◐ | Cache distribuído por hashring + dispatch, réplicas de leitura, query planner (2026). Materialize pago. Embutido só para Check |
| 10. Auditoria | ◐ | Watch como feed de mudanças. "Request & Response Audit Logging" só em planos pagos (página de preços) |
| 11. Experiência e administração | ◐ | Playground (WASM), `zed` CLI, LSP, clientes oficiais. Sem UI administrativa no OSS (#933 aberta) |

## Diferenciais / exclusivos
- **ZedTokens + `at_least_as_fresh` / `at_exact_snapshot` / `fully_consistent`**: a solução mais completa do mercado aberto para o new enemy problem.
- **Resposta condicional de caveat** com os campos de contexto que faltam: o chamador pode completar e reavaliar.
- **Watch API** com retomada a partir de um ZedToken (base para sincronização downstream e caches).
- **Relationship integrity** (assinatura HMAC de relações no CockroachDB) e **expiração** nativa de relações.
- Suporte a **Spanner/CockroachDB multi-região**.
- **Materialize** (pago): pré-computa permissões transitivas e as transmite para índices de busca (sem suporte a caveats, wildcards e expiração no caminho materializado; com atraso).

## O que a comunidade elogia (com evidência)
- Rigor de consistência; vista como a mais "Zanzibar-fiel" (HN 28709886, 197 pontos; comparativos de terceiros).
- Ferramentas de validação e teste de schema (playground, `zed validate`, actions de CI).
- Transparência técnica: blogs sobre query planner com benchmarks (fornecedor).

## Do que a comunidade se queixa (com evidência)
- **Sincronização / escrita dupla**: no lançamento, usuários perguntam como manter Postgres e SpiceDB em sincronia (cascata de deleções); a resposta do fundador é "very application specific" (HN 28709886).
- **Listagem em escala**: a doc diz que LookupResources "can cause performance problems when more than 10k results"; issue #1317 (LookupResources "through objects") está entre as mais votadas e aberta.
- **Expressividade**: setas aninhadas arbitrárias (#15, aberta desde o início), relações inversas (#356), multi-tenancy (#204), multi-test validation (#1437, 47 comentários).
- **Embutir como biblioteca** (#205): mantenedores resistiram ("perde valor"), depois ofereceram exemplo com bufconn; o pacote `pkg/embedded` (só Check) veio depois.
- **Operação**: CockroachDB recomendado, mas com custo operacional próprio; `fully_consistent` **não** garante read-after-write no CockroachDB (skew de até 500 ms, conforme a doc).

## Histórico de segurança
18 advisories (2022–2026), na maioria de severidade baixa, concentrados em **corretude de LookupResources, caveats e cache**:
- GHSA-7p8f-8hjm-wm92 (2022, alta): lookups ignoravam wildcards em interseções/exclusões.
- CVE-2023-35930 (GHSA-m54h-5x5f-5m6r), CVE-2025-65111 (GHSA-9m7r-g8hg-x3vr), GHSA-j85q-46hg-36p2: **resultados parciais** em LookupResources/LookupSubjects.
- GHSA-h3m7-rqc4-7h9p (2024, alta): overflow de inteiro no dispatch fazia perder elementos ou entrar em pânico.
- GHSA-4vrg-r928-h5vv (2026): caveat resultava em permissão **incondicional**. GHSA-mqcf-gqvg-rmhm (2026): reuso indevido de cache com listas aninhadas em caveats. GHSA-5784-6qcr-48fq (2026): `in_cidr` não casava IPv4 mapeado em IPv6.
- GHSA-vhvq-fv9f-wh4q (2026): cursor de LookupResources adulterado derrubava o processo (pânico).
- Vazamento de segredos: URI do datastore em logs (2023, 2026), em traces (2026), flags expostas via porta de métricas (2023).
- GHSA-pm3x-jrhh-qcr7 (2025): WriteRelationships falhava silenciosamente com payload grande.
- **Lições**: (1) cursores e tokens opacos precisam de autenticação (MAC) e de parsing sem pânico; (2) chave de cache deve incluir todo o contexto de caveat normalizado; (3) segredos nunca em logs, traces ou endpoints de métricas (redação central); (4) Lookup e Check devem ser testados um contra o outro; (5) falha de escrita nunca silenciosa.

## Arquitetura e decisões notáveis
- Servidor stateless com **dispatch** de subproblemas para o nó dono via hashring consistente: cache distribuído.
- Snapshots MVCC no datastore; janela de GC (`--datastore-gc-window`) limita `at_exact_snapshot` e Watch histórico. No Postgres, `track_commit_timestamp=on` é obrigatório para Watch.
- Quantização de revisões em `minimize_latency` para aumentar o acerto do cache.
- Query planner (2026) com "advisors" estático e por contagem, que substitui gradualmente o dispatch ingênuo.
- Schema único por permissions system; mudanças validadas contra relações existentes.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: **não nativa** (#204 aberta); prefixos de tipo ou um permissions system por tenant (Cloud).
- Extensibilidade: interface de datastore em Go; caveats CEL; middlewares.
- Configuração por flags/env (`SPICEDB_*`), operador Kubernetes.
- Bancos: CockroachDB (recomendado para multi-região e alto throughput), Spanner (GCP), PostgreSQL (região única), MySQL (não recomendado), memdb (dev/teste).
- Soberania: self-hosted total; Cloud em regiões AuthZed; Dedicated com rede privada.

## (Somente AuthZ) Modelos, linguagem de políticas, consistência, listagem, sincronização de dados, explicabilidade, testes de políticas, avaliação local, desempenho publicado, AuthZEN
- **Linguagem**: schema `definition`/`relation`/`permission` com `+`, `&`, `-`, `->` (setas), caveats `caveat nome(params) { CEL }`, tipos `ipaddress`, `duration`, `timestamp`, `map`, `list`. No contexto do caveat, o valor gravado na relação prevalece sobre o da requisição.
- **Consistência**: escrita e schema `fully_consistent`; consultas `minimize_latency` por padrão. Recomendação oficial: **guardar o ZedToken na linha do recurso** (`text` ou `varchar(1024)`) e atualizá-lo quando o recurso ou o acesso muda (equivalente ao Content-Change Check do Zanzibar).
- **Listagem**: LookupResources com cursor (pré-filtro); acima de ~10 mil resultados, a recomendação é pós-filtrar com CheckBulk e depois Materialize (pago). LookupSubjects para "quem acessa".
- **Sincronização**: responsabilidade da aplicação (escrita direta, outbox, CDC). Watch permite fluxo SpiceDB→sistemas downstream. Materialize transmite permissões computadas para busca. Não há conector CDC oficial de entrada encontrado.
- **Explicabilidade**: `zed permission check --explain`, debug traces na resposta de Check.
- **Testes**: arquivo de validação YAML (schema, relações, assertions `assertTrue/False/Caveated`, expected relations), `serve-testing` (servidor de teste isolado por token), actions `action-spicedb` e `action-spicedb-validate`.
- **Avaliação local / embutível**: `pkg/embedded` roda o motor em processo, **apenas `Check`**, sempre totalmente consistente (sem lookup, escrita ou watch); alternativa com bufconn e `pkg/cmd`. WASM no playground. Wrappers comunitários em Python e .NET via cgo.
- **Desempenho publicado**: README cita "5 ms p95" com milhões de QPS e bilhões de relações (fornecedor); benchmarks do query planner (blog, 05/2026).
- **AuthZEN**: **não encontrado** (docs, issues, lista de PDPs do interop da OpenID).
- **Custos operacionais**: serviço com vários nós (o cache depende do cluster), banco distribuído recomendado (CockroachDB/Spanner), GC de revisões, migrações de datastore em upgrades. Cloud a partir de ~US$ 2/h (~US$ 1,4 mil/mês) por sistema; recursos de escala (Materialize) só pagos.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Nenhum fluxo de solicitação ou aprovação; relações com expiração ajudam apenas no acesso temporário depois da aprovação, feita fora.
- **Autorização sensível ao estado / transições declarativas**: **parcial**. O estado pode ser parâmetro de caveat (ex.: `status == "draft"`), mas não há estados, transições nem políticas por transição.

## Implicações para o projeto
1. **Adotar**: tokens de consistência no estilo ZedToken (4 níveis) e a recomendação de guardá-los junto ao recurso. É o padrão de referência para o new enemy problem.
2. **Adotar**: resultado tri-estado (permitido, negado, condicional + atributos faltantes). Encaixa em AuthZEN e em "obrigações".
3. **Adotar**: feed de mudanças retomável por token (Watch) como base de auditoria e de sincronização downstream.
4. **Oportunidade**: listagem eficiente em grandes volumes **sem paywall**; o mercado cobra por isso (Materialize). Avaliar materialização incremental no próprio motor.
5. **Oportunidade**: multi-tenancy nativa (lacuna #204) e AuthZEN (ausente).
6. **Lição de segurança**: tokens e cursores autenticados, redação de segredos centralizada, chave de cache completa, testes cruzados Check×Lookup.
7. **Evitar**: depender de banco distribuído caro para a consistência básica; o Postgres com snapshots basta no estágio inicial (premissa 17).

## Fontes
- https://github.com/authzed/spicedb (README) — 2026-11-03
- https://proxy.golang.org/github.com/authzed/spicedb/@latest (v1.57.0, 2026-10-20) — 2026-11-03
- https://github.com/authzed/docs: `app/spicedb/concepts/{consistency,caveats,datastores,querying-data,watch,expiring-relationships}/page.mdx`, `app/spicedb/modeling/validation-testing-debugging/page.mdx` (via raw; WebFetch de authzed.com bloqueado) — 2026-11-03
- https://raw.githubusercontent.com/authzed/spicedb/main/pkg/embedded/README.md — 2026-11-03
- https://authzed.com/pricing (curl) — 2026-11-03 — fornecedor
- https://authzed.com/blog/authzed-materialize-generally-available (10/09/2026) e https://authzed.com/blog/query-planner-early-benchmarks (20/05/2026) — 2026-11-03 — fornecedor
- https://github.com/authzed/spicedb/security/advisories (páginas 1–2) — 2026-11-03
- https://github.com/authzed/spicedb/issues?q=sort:reactions (#15, #204, #1317, #1437, #205, #933) — 2026-11-03
- https://github.com/authzed/spicedb/issues/205 — 2026-11-03
- https://news.ycombinator.com/item?id=28709886 (fundadores AuthZed respondem) — 2026-11-03
- https://news.ycombinator.com/item?id=45661547 (comentário de jschorr/AuthZed e da Oso; ambos com viés) — 2026-11-03
- https://raw.githubusercontent.com/openid/authzen/main/interop/authzen-todo-backend/src/pdps.json (SpiceDB ausente) — 2026-11-03
- https://osv.dev/vulnerability/CVE-2025-65111 (via busca) — 2026-11-03
- https://pypi.org/project/spicedb-embedded/ e https://www.nuget.org/packages/SpiceDbEmbedded (comunitários; via busca) — 2026-11-03
