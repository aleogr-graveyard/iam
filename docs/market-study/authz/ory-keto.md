# Ory Keto (Ory Permissions)

| Campo | Valor |
|---|---|
| Categoria | Motor relacional estilo Zanzibar (ReBAC), servidor; parte do ecossistema Ory (Kratos, Hydra, Oathkeeper). Versão gerenciada: "Ory Permissions" na Ory Network |
| Site / repositório | https://www.ory.com/docs/keto · https://github.com/ory/keto |
| Licença e modelo de negócio | Apache-2.0 no núcleo. **Ory Enterprise License (OEL)**: recursos extras, builds privados, **patches de CVE com SLA** e multi-tenancy avançada só no registro Docker privado. O README diz que o OSS serve para "unimportant workloads without SLAs". Ory Network: US$ 0,00009 por checagem/mês no plano Production (~US$ 90 por milhão) |
| Stack | Go; PostgreSQL, MySQL, CockroachDB (SQLite em dev/testes). gRPC + REST. OPL (subconjunto de TypeScript) |
| Maturidade | Projeto desde 2018 (motor ACP/OPA), reescrito como Zanzibar em 04/2021. Releases: v0.14.0 (03/2025), v25.4.0 (11/2025), v26.2.0 (03/2026). ~5,4 mil estrelas, ~390 forks. O código OSS é espelhado de um monorepo privado (commits "GitOrigin-RevId", Copybara). Adotantes citados no README: OpenAI, Klarna, Cisco, Mistral (para o ecossistema Ory, não necessariamente Keto) |
| Acesso | 2026-11-03 |

## Resumo
- Foi o primeiro Zanzibar open source anunciado (HN 2021), mas o mais básico do grupo: só ReBAC, com **OPL** (sintaxe TypeScript) para reescritas de permissão.
- **Sem tokens de consistência**: o campo `snaptoken` existe no proto, mas comentado ("<!-- -->") e não implementado.
- **Sem ABAC/condições**: a issue #319 foi fechada pelo bot de "stale" em 07/2026 como "not planned".
- Listagem limitada: a "List API" devolve **tuplas armazenadas** (consulta parcial), não permissões computadas; há Expand para "quem tem acesso".
- Modelo open core com SLA de segurança só na OEL: risco para quem self-hospeda.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | ReBAC e RBAC via relações e OPL (`includes`, `traverse`, `&&`, `\|\|`, `!`). Sem ABAC nem PBAC |
| 2. Decisões | ◐ | Check, batch check (v0.14.0), Expand, List de tuplas. Sem listagem de objetos/sujeitos por permissão computada; sem explicação além do Expand |
| 3. Políticas | ◐ | OPL editável na Ory Console; sem ferramenta de testes de políticas encontrada (procurado em docs e repositório) |
| 4. Dados | ◐ | Tuplas e subject sets; mapeamento UUID↔nome; inserção/remoção em lote. Sem snaptoken, sem Watch encontrado |
| 5. Acesso dinâmico | ○ | Sem expiração/TTL verificada (issue #346 fechada, implementação não verificada), sem solicitações ou aprovações |
| 6. Estados e transições | ○ | Ausente |
| 7. Padrões | ○ | AuthZEN não encontrado (docs, interop OpenID) |
| 8. Máquinas e agentes | ◐ | Sujeitos arbitrários; Ory oferece M2M tokens e "Agent IAM" no ecossistema, não no Keto |
| 9. Desempenho e resiliência | ◐ | README promete checagens abaixo de 10 ms e escala horizontal; `max_depth` por requisição. Não é biblioteca |
| 10. Auditoria | ? | Não encontrado no OSS |
| 11. Experiência e administração | ◐ | Ory Console (Network) com editor OPL; CLI e SDKs gerados. Sem UI no OSS |

## Diferenciais / exclusivos
- **OPL com sintaxe TypeScript**: familiar e com tooling de editor; posicionada contra Rego e Casbin.
- Integração nativa com o restante do Ory (Kratos para identidade, Oathkeeper como PEP proxy).
- Preço por checagem publicado na Ory Network (raro no mercado).

## O que a comunidade elogia (com evidência)
- Discussão de design aberta e de alto nível (issue #47 "Moving forward with ORY Keto", 50 👍, citada no HN 26738344).
- Uso bem-sucedido em B2B SaaS self-hosted (HN 36470943, depoimento de usuário).
- OPL considerada mais acessível que DSLs proprietárias (doc oficial; fornecedor).

## Do que a comunidade se queixa (com evidência)
- **Consistência**: no lançamento, "Keto just does not" implementa as garantias de consistência do Zanzibar (HN 26738344). Continua sem snaptoken em 2026 (proto).
- **Funcionalidades faltantes**: o fundador da AuthZed diz que Keto "is missing lots of the core functionality" (HN 28709886; concorrente). Issues abertas: OR dentro de `traverse` (#1250), comparar `this` com `ctx.subject` (#1204).
- **ABAC**: #319 encerrada por inatividade; a sugestão era usar OPA junto.
- **Embutir**: "Keto can not be consumed as a dependency" (#64, 11 👍).
- **Open core**: CVE fixes garantidos apenas com OEL (README). A cadência pública é irregular (gap de 13 meses entre 0.13 e 0.14).

## Histórico de segurança
- **CVE-2026-33505** (GHSA-c38g-mx2c-9wf2, 03/2026, alta, CVSS 7.2, CWE-89): **SQL injection via token de paginação forjado** em GetRelationships. Os tokens eram cifrados com `secrets.pagination`; se não configurado, usava **segredo padrão embutido no código**, público, o que permitia forjar tokens. Afeta v25.4.0; corrigido em v26.2.0.
- Changelog cita atualizações de dependências por GHSA-p77j-4mvh-x3m3 (marcado "critical") e GHSA-7h2j-956f-4vf2.
- **Lições**: (1) nunca ter segredo padrão; falhar na inicialização sem segredo ou gerar um aleatório persistido (alertar em vez de bloquear, mas nunca usar valor público); (2) tokens opacos de paginação devem ser autenticados (MAC/AEAD) e o conteúdo **nunca** interpolado em SQL; (3) "alertas de configuração" (brief 4.6) devem cobrir segredos ausentes.

## Arquitetura e decisões notáveis
- Servidor stateless com APIs de leitura e escrita em portas distintas; persistência SQL via `ory/x` (pop/migrations).
- Isolamento por `nid` (network ID) nas tabelas: a base da multi-tenancy da Ory Network ("Force SQL operator precedence... to ensure nid isolation" no changelog).
- Avaliação de check com profundidade máxima global e por requisição; índices de busca reversa (v0.9).
- Desenvolvimento em monorepo privado espelhado: contribuições externas e transparência de roadmap reduzidas.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: namespaces no OSS; isolamento por projeto/`nid` na Ory Network; "advanced multi-tenancy" vendida via OEL.
- Extensibilidade: baixa; código em `internal/`, não consumível como biblioteca (#64).
- Configuração: YAML/env, padrão Ory.
- Bancos: PostgreSQL, MySQL, CockroachDB; SQLite para dev.
- Soberania: self-hosted ou Ory Network com localidade de dados "GDPR friendly" (README).

## (Somente AuthZ) Modelos, linguagem de políticas, consistência, listagem, sincronização de dados, explicabilidade, testes de políticas, avaliação local, desempenho publicado, AuthZEN
- **Linguagem**: OPL (classes = namespaces, `related` = relações, `permits` = funções booleanas). Sem condições ou atributos.
- **Consistência**: sem zookies ou snaptokens (campo comentado no proto `check_service.proto`). Avaliação sobre o estado mais recente; o new enemy problem não é tratado explicitamente.
- **Listagem**: List API filtra tuplas por consulta parcial (objeto, relação, sujeito) sem resolver reescritas; Expand monta a árvore de sujeitos. Não equivale a ListObjects/LookupResources.
- **Sincronização**: responsabilidade da aplicação; nenhum Watch, changelog ou CDC encontrado.
- **Explicabilidade**: Expand.
- **Testes**: não encontrado.
- **Avaliação local / embutível**: não (só servidor).
- **Desempenho publicado**: só a alegação "sub-10ms" do README; sem benchmark público encontrado.
- **AuthZEN**: não encontrado.
- **Custos operacionais**: serviço + SQL; baixo custo de infraestrutura, mas o custo de risco é alto no OSS (sem SLA de CVE). Na Network, ~US$ 90 por milhão de checagens (Production), além do plano base (US$ 770/ano no plano de entrada exibido).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**.
- **Autorização sensível ao estado / transições declarativas**: **não**. Sem condições, o estado só poderia ser modelado como tuplas.

## Implicações para o projeto
1. **Evitar**: segredo padrão embutido e tokens de paginação decodificados para SQL (CVE-2026-33505). Entra no modelo de ameaças.
2. **Evitar**: open core em que patches de segurança dependem de licença paga; conflita com as premissas 2 e 11 (a monetização deve vir de serviço ou recursos, não de segurança).
3. **Oportunidade**: a linguagem com sintaxe familiar (TypeScript) foi bem recebida; avaliar a DX da nossa linguagem de políticas por esse critério.
4. **Evitar**: lançar com a promessa "Zanzibar" sem tokens de consistência; a comunidade cobrou desde o dia 1.
5. **Adotar**: isolamento de tenant por coluna obrigatória em toda consulta (`nid`), com testes que garantam a precedência de operadores (bug real no changelog).

## Fontes
- https://github.com/ory/keto (README: OEL, deployment, bancos) e https://raw.githubusercontent.com/ory/keto/master/README.md — 2026-11-03
- https://raw.githubusercontent.com/ory/keto/master/CHANGELOG.md — 2026-11-03
- https://raw.githubusercontent.com/ory/keto/master/proto/ory/keto/relation_tuples/v1alpha2/check_service.proto (snaptoken comentado) — 2026-11-03
- https://github.com/ory/docs: `docs/keto/index.mdx`, `reference/ory-permission-language.mdx`, `guides/list-api-display-objects.mdx`, `concepts/25_api-overview.mdx` (via raw) — 2026-11-03
- https://github.com/ory/keto/security/advisories/GHSA-c38g-mx2c-9wf2 — 2026-11-03
- https://github.com/ory/keto/issues?q=sort:reactions (#47, #64, #319, #346, #1204, #1250) e https://github.com/ory/keto/issues/319 — 2026-11-03
- https://www.ory.com/pricing (curl; preço por checagem) — 2026-11-03 — fornecedor
- https://news.ycombinator.com/item?id=26738344 (mantenedores Ory respondem), https://news.ycombinator.com/item?id=28709886 (viés: fundadores da AuthZed, concorrente), https://news.ycombinator.com/item?id=36470943 — 2026-11-03
- https://proxy.golang.org/github.com/ory/keto/@latest (v0.14.0 como último semver Go) — 2026-11-03
- https://raw.githubusercontent.com/openid/authzen/main/interop/authzen-todo-backend/src/pdps.json (Keto ausente) — 2026-11-03
