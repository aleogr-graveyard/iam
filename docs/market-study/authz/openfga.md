# OpenFGA (inclui Auth0 FGA)

| Campo | Valor |
|---|---|
| Categoria | Motor relacional estilo Zanzibar (ReBAC), servidor; Auth0 FGA é a versão hospedada pela Okta/Auth0 |
| Site / repositório | https://openfga.dev · https://github.com/openfga/openfga · hospedado: https://auth0.com/fine-grained-authorization |
| Licença e modelo de negócio | Apache-2.0, sem mudanças de licença. Projeto CNCF (Sandbox em 2022, **Incubating desde 28/10/2025**). Monetização indireta pela Okta (Auth0 FGA, preço sob consulta, "Public Cloud e Private Cloud AWS") |
| Stack | Go; PostgreSQL 14+ e MySQL 8 (produção), SQLite (beta), memória (dev). HTTP + gRPC. CEL nas condições. SDKs Java, Node.js, Go, Python, .NET |
| Maturidade | v0.1.0 em 06/2022, v1.0.0 em 04/2023, **v1.22.1 em 27/10/2026** (proxy.golang.org). ~6,1 mil estrelas, ~550 forks. Em produção no Auth0 FGA desde 12/2021. Adotantes citados: Auth0, Grafana Labs, Canonical, Docker, Read.AI; CNCF fala em 37 empresas em produção |
| Acesso | 2026-11-03 |

## Resumo
- Implementação Zanzibar mais próxima de um "padrão aberto neutro": governança CNCF, Apache-2.0 e a única do grupo com AuthZEN implementado (experimental).
- Modelo em DSL própria (`.fga`) com tipos, relações, união, interseção e exclusão (`but not`), mais **condições CEL** em tuplas (ABAC sobre ReBAC).
- Ponto fraco estrutural: **não tem zookies**. Só há `MINIMIZE_LATENCY` e `HIGHER_CONSISTENCY`, que pula o cache.
- ListObjects e ListUsers têm limite de tempo (3 s) e de resultados (1000), sem paginação. Em 2026 o ListObjects foi reescrito como pipeline em streaming.
- Pode rodar **embutido em processo Go** (`pkg/server`, como no Grafana), mas a escrita não participa da transação da aplicação: a escrita dupla continua sendo problema da aplicação.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC e ReBAC nativos; ABAC via condições CEL em tuplas e via contexto da requisição. Sem PBAC (não há políticas declarativas fora do grafo) |
| 2. Decisões | ◐ | Check, BatchCheck, ListObjects, StreamedListObjects, ListUsers, Expand, Read. Não há "explain": só o Expand (árvore de usersets) |
| 3. Políticas | ◐ | DSL própria; modelos imutáveis com ID (versionamento implícito, fixável por requisição); módulos (modular models); testes `fga model test` em `.fga.yaml`. Sem simulação nem implantação gradual nativas |
| 4. Dados | ◐ | Tuplas, condições com contexto persistido, tuplas contextuais (efêmeras, por requisição), ReadChanges (feed ordenado de mudanças). Sem tokens de consistência |
| 5. Acesso dinâmico | ◐ | Acesso temporário via condição (ex.: `grant_time + duration`). Sem solicitações/aprovações, sem M de N, sem obrigações |
| 6. Estados e transições | ○ | Estado só pode entrar como atributo de condição; sem transições declarativas |
| 7. Padrões | ◐ | AuthZEN 1.0 experimental (avaliação, lote, buscas de sujeito/recurso/ação, descoberta). Sem RAR/UMA |
| 8. Máquinas e agentes | ◐ | Sujeitos de qualquer tipo (serviços, agentes); Gravitee usa AuthZEN→OpenFGA para MCP (preview). Nada específico de MCP no núcleo |
| 9. Desempenho e resiliência | ◐ | Caches de subproblemas e de iteradores, cache controller, planejador adaptativo de estratégias (2026), réplica de leitura (Postgres; MySQL desde 1.22). Embutível em Go. Sem avaliação local em SDK |
| 10. Auditoria | ◐ | ReadChanges registra escritas de tuplas; log de decisões não encontrado no OSS |
| 11. Experiência e administração | ◐ | Playground, CLI, extensão VS Code, 5 SDKs oficiais. Sem UI administrativa de tuplas no OSS (Auth0 FGA tem dashboard) |

## Diferenciais / exclusivos
- **Único do grupo com AuthZEN** implementado e presente no interop da OpenID (pdps.json lista OpenFGA; SpiceDB, Permify e Keto ausentes).
- **Tuplas contextuais**: relações efêmeras enviadas na requisição, úteis para dados que não se quer sincronizar (ex.: grupos vindos do token).
- **Modelos imutáveis com ID**: cada escrita de modelo gera versão nova; requisições podem fixar `authorization_model_id`.
- **StreamedListObjects** sem limite de resultados (só prazo) e pipeline de ListObjects com poda de arestas.
- Governança neutra (CNCF Incubating), com mantenedores de Okta, Grafana e GitPod.

## O que a comunidade elogia (com evidência)
- Documentação de modelagem e DSL legível; SDKs oficiais em 5 linguagens (README).
- Adoção por empresas conhecidas e embutimento real em produto (Grafana "Zanzana", modo embutido em processo).
- Investimento visível em desempenho: planejador "self-tuning" para P99 e novo algoritmo de ListObjects (blogs Auth0, 2026; fornecedor).

## Do que a comunidade se queixa (com evidência)
- **Sem zookies**: citado no HN como diferença central em relação ao SpiceDB (item 41311489). A própria doc lista zookies como "future work".
- **ListObjects**: resultados truncados por prazo ou limite sem aviso (issue #1961 aberta); sem paginação (#2828 aberta); lento em modelos com `and`/`but not` (aviso na doc; #1607). Um usuário multi-tenant teme que uma busca afete os demais tenants (community #104).
- **Sincronização**: "keeping everything in sync all the time" ao reescrever OpenFGA em Postgres puro (HN 45661547; comentário da Oso, concorrente).
- **Embutido sem transação**: issue #3302 (09/2026) pede que o servidor embutido participe da transação do host; hoje é preciso um outbox. Sem resposta de mantenedor.
- Issues mais votadas são de ergonomia operacional: StoreID predefinido (#1132), playground local (#338), deadline_exceeded e falta de conexões no banco (#2422).

## Histórico de segurança
29 advisories GHSA de 10/2022 a 09/2026, quase todos da classe **"Authorization Bypass / Improper Policy Enforcement"** em casos de borda do avaliador:
- CVE-2024-42473 (GHSA-3f6g-m4hr-59h8, 2024, alta): bypass no Check com `but not` + `from` + userset (v1.5.7–1.5.8).
- CVE-2025-48371 (GHSA-c72g-53hw-82q7, 2025, moderada): tuplas contextuais com userset em relação com acesso público tipado.
- GHSA-5278-rrxc-mgf7 (2026, moderada): **fail-open**. No pipeline de StreamedListObjects, um erro no lado subtraído da exclusão (contexto ausente, falha do banco) deixava passar objetos excluídos (CWE-636).
- GHSA-h7w8-xr72-cv4r (2026): ListUsers retornava usuário que deveria ser excluído (exclusão aninhada com wildcard `user:*`).
- CVE-2026-55170 (GHSA-cf98-j28v-49v6, 2026, baixa): colação case-insensitive no MySQL. A correção exige janela de manutenção.
- Outras classes: DoS por definições circulares (2023), exposição da chave do Playground (2026), autenticação imprópria (2026), host-header poisoning na descoberta AuthZEN (corrigido antes de advisory).
- **Lições**: (1) exclusão + wildcard + userset + contexto é a área mais frágil e precisa de testes diferenciais e de propriedade entre Check, ListObjects e ListUsers (as três devem concordar); (2) erro em subárvore negativa nunca pode virar permissão (fail closed também no meio da avaliação); (3) diferenças de colação entre bancos são falha de autorização; (4) metadados de descoberta não podem ser derivados de cabeçalhos da requisição.

## Arquitetura e decisões notáveis
- Servidor stateless + banco relacional; tuplas `objeto#relação@usuário[condição]`.
- Avaliação recursiva com limites de profundidade e de largura (`resolve-node-limit`, `resolve-node-breadth-limit`) e limites de concorrência por API. Proteção explícita contra consultas caras.
- Planejador de estratégias de Check por estatística em tempo de execução (Thompson sampling, `internal/planner`).
- Cache recomendado: poucos servidores grandes para aumentar o acerto do cache (doc "Running in Production").
- Banco recomendado **exclusivo** do OpenFGA, migrado por `openfga migrate`.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy por **stores** (modelo e tuplas isolados por store); Grafana usa isso em SaaS multi-tenant.
- Extensibilidade: interface de storage em Go (adaptador SQLite contribuído pelo Grafana); experimentais por flag.
- Configuração por flags, env (`OPENFGA_*`) e YAML.
- Bancos: PostgreSQL, MySQL, SQLite (beta), memória.
- Soberania: self-hosted total; Auth0 FGA em nuvem pública ou privada AWS (regiões não verificadas).

## (Somente AuthZ) Modelos, linguagem de políticas, consistência, listagem, sincronização de dados, explicabilidade, testes de políticas, avaliação local, desempenho publicado, AuthZEN
- **Linguagem**: DSL `.fga` (convertida para JSON); condições em CEL com parâmetros tipados (timestamp, duration, ipaddress etc.). Na fusão de contextos, o contexto persistido na tupla **prevalece** sobre o da requisição.
- **Consistência**: `MINIMIZE_LATENCY` (padrão, usa cache) e `HIGHER_CONSISTENCY` (sem cache, leitura forte; na réplica, vai ao primário). Sem zookies e, portanto, sem proteção barata contra o **new enemy problem**. A doc sugere guardar `modified_date` no recurso e escolher o modo em tempo de execução.
- **Listagem**: ListObjects (padrão: 3 s e 1000 resultados), StreamedListObjects (sem limite de quantidade), ListUsers (3 s e 1000). A doc recomenda "search then BatchCheck" para paginação.
- **Sincronização**: responsabilidade da aplicação; ReadChanges permite consumidores downstream. Sem CDC oficial (há projeto comunitário Keycloak→OpenFGA, HN 47399437). Embutido: escrita fora da transação do host (#3302).
- **Explicabilidade**: Expand. Sem trilha de "por quê".
- **Testes**: `fga model test` com tuplas, checks e list_objects esperados em YAML (CLI).
- **Avaliação local / embutível**: `pkg/server.NewServerWithOpts` em processo (Grafana: "in-proc gRPC"). Não há avaliação local em SDK.
- **Desempenho publicado**: posts de engenharia (planejador, pipeline) sem benchmark reprodutível independente encontrado.
- **AuthZEN**: experimental (`--experimentals=authzen`) desde v1.12/1.13 (03/2026). Mapeia Evaluation→Check, Subject Search→ListUsers, Resource Search→ListObjects; Action Search avalia todas as relações do tipo. Desvios: **sem paginação** nas buscas, sem `signed_metadata` nem URNs de capacidades, cabeçalho proprietário `Openfga-Authorization-Model-Id`, sem tuplas contextuais. Propriedades do sujeito, recurso e ação viram contexto com prefixos `subject_`, `resource_` e `action_`.
- **Custos operacionais**: serviço + Postgres/MySQL dedicado; ajuste fino de pool de conexões, limites de concorrência e caches; upgrades com migrações (algumas com janela de manutenção). Auth0 FGA: preço não publicado (vendas). Comparativos de terceiros chamam produção do FGA de contrato Enterprise (Clerk, concorrente).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Nenhuma API de solicitação ou aprovação; o modelo só expressa relações estáticas e condições.
- **Autorização sensível ao estado / transições declarativas**: **parcial**. O estado pode entrar como parâmetro de condição CEL (ex.: `status == "draft"`), mas não há conceito de estado, transição ou política por transição.

## Implicações para o projeto
1. **Adotar**: a forma do AuthZEN do OpenFGA (store/tenant no caminho, propriedades→contexto com prefixo) é referência concreta. Superar os desvios: paginação nas buscas e metadados com capacidades.
2. **Adotar**: limites explícitos de profundidade, largura, concorrência e prazo por API, configuráveis por tenant (premissa 18), sinalizando **truncamento** na resposta (lacuna #1961).
3. **Evitar**: ausência de token de consistência. Desenhar tokens tipo ZedToken desde o início.
4. **Oportunidade**: motor embutido que **participa da transação da aplicação** (#3302 aberta), resolvendo a escrita dupla no modo monólito modular.
5. **Lição de segurança**: testes diferenciais e de propriedade entre Check, List e Search, e fail closed em erro de subárvore negativa.
6. **Adotar**: modelos imutáveis com ID e fixação por requisição como base de versionamento e implantação gradual.

## Fontes
- https://github.com/openfga/openfga (README, estrelas, bancos, adotantes) — 2026-11-03
- https://raw.githubusercontent.com/openfga/openfga/main/CHANGELOG.md — 2026-11-03
- https://proxy.golang.org/github.com/openfga/openfga/@latest (v1.22.1, 2026-10-28) — 2026-11-03
- https://github.com/openfga/openfga.dev, docs `interacting/authzen.mdx`, `interacting/consistency.mdx`, `modeling/conditions.mdx`, `interacting/relationship-queries.mdx`, `getting-started/perform-list-objects.mdx`, `interacting/read-tuple-changes.mdx`, `best-practices/running-in-production.mdx` (via raw.githubusercontent; openfga.dev bloqueado nesta sessão) — 2026-11-03
- https://www.cncf.io/blog/2025/11/11/openfga-becomes-a-cncf-incubating-project/ (via busca) — 2026-11-03
- https://github.com/openfga/openfga/security/advisories (páginas 1–3) e advisories GHSA-3f6g-m4hr-59h8, GHSA-c72g-53hw-82q7, GHSA-5278-rrxc-mgf7, GHSA-h7w8-xr72-cv4r — 2026-11-03
- https://github.com/openfga/openfga/issues?q=sort:reactions (issues #1132, #1961, #2828, #3302, #2422, #1607) — 2026-11-03
- https://github.com/openfga/openfga/issues/3302 — 2026-11-03
- https://github.com/orgs/openfga/community/issues/104 (via busca) — 2026-11-03
- https://github.com/grafana/grafana/tree/main/pkg/services/authz/zanzana (README: modo embutido) — 2026-11-03
- https://raw.githubusercontent.com/openid/authzen/main/interop/authzen-todo-backend/src/pdps.json — 2026-11-03
- https://raw.githubusercontent.com/openfga/cli/main/README.md (`fga model test`) — 2026-11-03
- https://news.ycombinator.com/item?id=45661547 (viés: comentário da Oso, concorrente), https://news.ycombinator.com/item?id=41311489 (viés: fundadores da Permify e ex-Warrant) — 2026-11-03
- https://auth0.com/blog/self-tuning-strategy-planner-openfga/ e https://auth0.com/blog/openfga-improved-listobjects-algorithm/ (fornecedor; via HN) — 2026-11-03
- https://auth0.com/pricing e https://auth0.com/fine-grained-authorization (sem preço público de FGA) — 2026-11-03
- https://documentation.gravitee.io/am/guides/authorization-engines/authzen (via busca; terceiro) — 2026-11-03
