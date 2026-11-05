# Permify (FusionAuth FGA by Permify)

| Campo | Valor |
|---|---|
| Categoria | Motor relacional estilo Zanzibar com ABAC (atributos + regras CEL), servidor. Desde 11/2025 é "a FusionAuth company" (produto de AuthZ irmão de um CIAM) |
| Site / repositório | https://permify.co · https://github.com/Permify/permify · docs no repositório (`/docs`, Mintlify) |
| Licença e modelo de negócio | **Mudou de Apache-2.0 para AGPL-3.0 em 12/05/2025** (commit 26912d9 "docs: update license"). Community Edition self-hosted; Permify Cloud com recursos premium (painéis de observabilidade, **sincronização de dados por CDC**), SOC 2, SLAs. Preços não verificados (permify.co/pricing não acessado) |
| Stack | Go; PostgreSQL 13.8+ (produção, com pgcat recomendado para múltiplas réplicas), memória (dev). REST (3476) + gRPC (3478). cel-go, ristretto (cache), pgx |
| Maturidade | Desde 07/2022; v1.0 em 08/2024; **v1.7.5 em 09/10/2026** (proxy.golang.org). ~6 mil estrelas, ~330 forks. Sem adotantes nomeados no README. Adquirida pela FusionAuth em 20/11/2025 |
| Acesso | 2026-11-03 |

## Resumo
- Zanzibar "orientado a desenvolvedor": DSL com `entity`, `relation`, `permission`, mais **atributos tipados** e **regras CEL**, o ABAC mais integrado ao grafo do grupo.
- **Multi-tenancy nativa** (`tenant_id` em todas as APIs) e **versões de schema** selecionáveis por requisição.
- **Snap tokens** (equivalente a zookies) baseados em ID de transação do Postgres (MVCC); cache com chave que inclui snapshot e versão do schema.
- Listagem: LookupEntity (com stream e cursor), LookupSubject e SubjectPermission (todas as ações de um sujeito num recurso).
- Pontos de atenção: AGPL (afeta embutir ou modificar), só Postgres, CDC só na nuvem, futuro sob a FusionAuth em definição.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC, ReBAC e ABAC no mesmo schema (atributos + regras). Sem PBAC |
| 2. Decisões | ◐ | Check, bulk check, LookupEntity (+stream), LookupSubject, SubjectPermission, Expand. Explicador "por quê" pedido na issue #1368 (aberta) |
| 3. Políticas | ◐ | DSL própria; **schema versionado** (versão por requisição); `permify validate` com cenários e assertions; GitHub Action; partial write de schema. Sem rollout gradual explícito |
| 4. Dados | ● | Relações + atributos tipados, snap tokens, Watch API, **Data Bundles** (gatilhos declarativos que geram tuplas/atributos a partir de um evento), GC de snapshots. CDC (Debezium + Kafka) só no Cloud |
| 5. Acesso dinâmico | ◐ | Regras com contexto (horário, dia útil, limites numéricos). Sem expiração nativa de relações verificada; sem solicitações ou aprovações |
| 6. Estados e transições | ○ | Estado como atributo + regra; sem transições |
| 7. Padrões | ○ | AuthZEN não encontrado (docs, README, interop OpenID) |
| 8. Máquinas e agentes | ◐ | Guia "LLM authorization" (filtragem de RAG); sem MCP |
| 9. Desempenho e resiliência | ◐ | Cache em memória + **cache distribuído por hash consistente** entre instâncias; leitor/escritor separados. Não embutível (#2878 aberta) |
| 10. Auditoria | ? | Watch como feed de mudanças; log de decisões não encontrado no CE |
| 11. Experiência e administração | ◐ | Playground, SDKs (Go, Node, Python, Java...), API de tenants. Sem UI administrativa no CE |

## Diferenciais / exclusivos
- **Atributos de primeira classe** (`entity:id$attr|tipo:valor`) e regras CEL no schema: ABAC sem enviar todo o contexto na requisição.
- **Data Bundles**: declarar no servidor "quando `organization_created` acontecer, crie estas tuplas e atributos". Reduz a lógica de sincronização espalhada na aplicação.
- **Tenancy nativa** e **schema versionado por requisição**.
- **SubjectPermission**: devolve o mapa de todas as permissões de um sujeito em um recurso (útil para UI).
- Cache distribuído por hash consistente sem precisar de um cluster "dispatch" completo.

## O que a comunidade elogia (com evidência)
- Show HN 1.0 com 154 pontos (HN 41311489) e Show HN 2022 com 122 pontos: DSL legível, playground, multi-tenancy.
- ABAC integrado ao ReBAC (Show HN "Permify Now Supports ABAC", 2023).
- Documentação com muitos exemplos de modelagem (Google Docs, Notion, healthcare etc.; docs).

## Do que a comunidade se queixa (com evidência)
- **Sincronização / escrita dupla**: usuário pergunta como manter dados em sincronia "in the presence of errors". A resposta do fundador é escrever dentro da transação da aplicação e **compensar com delete** em caso de falha (HN 41311489). É uma escrita dupla, sem garantia.
- **"Crowded space"**: "How do you compare to keto, spicedb, topaz, warrant..." (HN 41311489); Zanzibar visto como "overkill" por parte da comunidade (HN 36470943).
- **Bancos**: CockroachDB/YugabyteDB pedidos e abertos (#605, #1111); só Postgres em produção.
- **Embutir**: "Embed Permify as an in-process Go library" (#2878, aberta); motores em `internal/`.
- **Licença**: troca para AGPL (2025) não comunicada com destaque (commit "docs: update license"); sites de avaliação alertam para revisão jurídica ao embutir (gappsy; terceiro).

## Histórico de segurança
- **Nenhum advisory publicado** na página de Security Advisories do GitHub (verificado em 2026-11-03).
- NVD não acessível nesta sessão (403); busca não encontrou CVEs do Permify.
- **Lição**: ausência de advisories não equivale a segurança. Motores com lógica comparável (OpenFGA, SpiceDB) acumulam dezenas de casos de corretude, e para nós a política de divulgação e o processo de advisories devem existir desde o início (premissa 5).

## Arquitetura e decisões notáveis
- Engine em dois passos: consulta o schema para achar as relações que compõem a ação e percorre o grafo com consultas concorrentes e curto-circuito em `or` (FAQ).
- Lookup reverso otimizado: primeiro acha as relações ligadas à ação e depois consulta só as tuplas daquele sujeito (doc LookupEntity).
- **MVCC no Postgres**: cada escrita cria snapshot (tabela `transactions`); snap token = ID de transação; GC remove versões antigas.
- Chave de cache `check_{tenant}_{schema_version}:{snapshot}:{request}`; sem TTL, despejo por memória (TinyLFU).
- Watch exige `track_commit_timestamp=on`.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: nativa (APIs de criar, listar e excluir tenant; schema e dados por tenant; tenant padrão `t1`).
- Extensibilidade: baixa como biblioteca; `pkg/development` (container em memória) serve playground e validação.
- Configuração: YAML/env; leitor e escritor separados; pgcat em modo sessão para múltiplas réplicas.
- Bancos: PostgreSQL (13.8+), memória.
- Soberania: CE em qualquer lugar; Cloud em AWS, GCP ou Azure com escolha de região (README).

## (Somente AuthZ) Modelos, linguagem de políticas, consistência, listagem, sincronização de dados, explicabilidade, testes de políticas, avaliação local, desempenho publicado, AuthZEN
- **Linguagem**: Permify Schema (`entity`, `relation`, `attribute`, `permission`, `rule nome(params) { CEL }`), sintaxe baseada em CEL.
- **Consistência**: snap token devolvido pela escrita; checagens com `snap_token` são avaliadas em snapshot pelo menos tão fresco quanto ele (padrão zookie). Sem token, usa a última transação (consulta extra ao banco). Não há níveis nomeados como no SpiceDB.
- **Listagem**: LookupEntity (array ou stream, com paginação por cursor desde 2024, #1253), LookupSubject, SubjectPermission. Benchmarks de lookup não encontrados.
- **Sincronização**: API de escrita; Data Bundles; Watch; CDC com Debezium + Kafka + Permify Sync Service **somente no Cloud**.
- **Explicabilidade**: Expand; explicador pedido (#1368).
- **Testes**: arquivo YAML com schema, relações, atributos e cenários com assertions (inclusive de lookup); `permify validate`; `permify-validate-action` no CI.
- **Avaliação local / embutível**: não suportado (#2878 aberta; Rust também pedido em #2875).
- **Desempenho publicado**: teste de carga no README: 1000 VUs, 10 mil RPS, 0% de falhas, média de 21,3 ms (fornecedor; sem metodologia detalhada); "até 10 ms" por checagem.
- **AuthZEN**: não encontrado.
- **Custos operacionais**: serviço + Postgres (+ pgcat em produção), GC de snapshots, cache por instância. Ao embutir ou modificar, a **AGPL** exige divulgar o código-fonte a usuários de rede.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**.
- **Autorização sensível ao estado / transições declarativas**: **parcial**. Atributo `status` + regra CEL permite "editar só em rascunho"; não há modelo de transições nem políticas por transição.

## Implicações para o projeto
1. **Adotar**: atributos de primeira classe junto das relações. ABAC sem o cliente enviar todo o contexto, com contexto da requisição como complemento.
2. **Adotar**: tenancy explícita em toda API e schema versionado por requisição (base para implantação gradual e rollback).
3. **Adotar (ideia)**: Data Bundles, ou seja, gatilhos declarativos de dados de autorização por evento. Combina com a "definição única" da seção 4.29 e reduz a escrita dupla espalhada.
4. **Evitar**: mudança de licença silenciosa; decidir a licença cedo e comunicar com clareza (premissa 11).
5. **Oportunidade**: oferecer o CDC/outbox de sincronização no OSS (no Permify é exclusivo da nuvem).
6. **Atenção**: AGPL impede estudar o código com liberdade para reaproveitar. Estudar apenas a documentação e o comportamento (regra do brief).

## Fontes
- https://github.com/Permify/permify (README: AGPL, Cloud vs CE, aquisição, teste de carga) — 2026-11-03
- https://raw.githubusercontent.com/Permify/permify/master/LICENSE (AGPL-3.0) e `.../1fba048/LICENSE` (Apache-2.0 antes) — 2026-11-03
- https://github.com/Permify/permify/commits/master/LICENSE (commit 26912d9 de 12/05/2025) — 2026-11-03
- https://permify.co/post/fusionauth-acquires-permify/ (20/11/2025; curl) — 2026-11-03 — fornecedor
- Docs no repositório (raw): `docs/operations/{snap-tokens,cache,bundle}.mdx`, `docs/getting-started/{sync-data,testing,modeling}.mdx`, `docs/integrations/data-sync.mdx`, `docs/api-reference/permission/{lookup-entity,subject-permission}.mdx`, `docs/setting-up/installation/database.mdx`, `docs/use-cases/{abac,multi-tenancy}.mdx`, `docs/permify-overview/faqs.mdx`, `docs/docs.json` — 2026-11-03
- https://raw.githubusercontent.com/Permify/permify/master/go.mod — 2026-11-03
- https://proxy.golang.org/github.com/!permify/permify/@latest (v1.7.5, 2026-10-09) — 2026-11-03
- https://github.com/Permify/permify/security/advisories (nenhum publicado) — 2026-11-03
- https://github.com/Permify/permify/issues?q=sort:reactions (#1368, #2878, #2875, #605, #1111, #1253) — 2026-11-03
- https://news.ycombinator.com/item?id=41311489 (viés: respostas de fundador), https://news.ycombinator.com/item?id=32096610, https://news.ycombinator.com/item?id=36470943 — 2026-11-03
- https://gappsy.com/tools/permify/ (via busca; agregador) — 2026-11-03
- https://raw.githubusercontent.com/openid/authzen/main/interop/authzen-todo-backend/src/pdps.json (Permify ausente) — 2026-11-03
