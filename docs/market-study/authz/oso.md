# Oso (Oso Cloud, Polar, Local Authorization, list filtering)

| Campo | Valor |
|---|---|
| Categoria | Plataforma de autorização como serviço (Oso Cloud) com linguagem própria (Polar); antes, biblioteca embutível open source |
| Site / repositório | https://www.osohq.com · docs: https://www.osohq.com/docs · biblioteca legada: https://github.com/osohq/oso (arquivada) |
| Licença e modelo de negócio | Oso Cloud é **proprietário** (SaaS, híbrido com fallback, self-hosted só em AWS). Biblioteca legada Apache-2.0, **arquivada em 22/10/2026** ("no longer maintained… no new releases or bug fixes"). Preço do Oso Cloud **não está mais público**: a página /pricing mostra apenas o novo produto "Oso for Agents" (US$ 0 / US$ 15 por usuário/mês / Enterprise). Planos citados na doc: Developer (grátis), Startup, Growth, Pro/Enterprise |
| Stack | Serviço em Rust (Rocket + SQLx, blog de 2022); fallback node em contêiner; CLI `oso-cloud`, Dev Server local; SDKs Node, Python, Ruby, Java, Go, .NET; extensão VS Code com LSP |
| Maturidade | Empresa de 2019+; Oso Cloud GA em 10/2022 (HN). Biblioteca legada: ~3,5 mil estrelas. **Em 06/10/2026 o Oso Cloud foi vendido à Tiny e a maior parte do time foi para a Temporal** (autorização de agentes). Futuro do produto incerto |
| Acesso | 2026-11-03 |

## Resumo
- Modelo "fatos + regras": dados de autorização são fatos (`has_role`, `has_relation`...) e a política em Polar (programação lógica) deriva permissões.
- Três modos de dados: fatos centralizados no Oso Cloud, fatos de contexto por requisição, e **Local Authorization** (Oso gera SQL que roda no banco da aplicação).
- Ferramentas de desenvolvimento fortes: testes dentro do Polar, Explain na UI, Dev Server local, Oso Migrate (paridade com o sistema legado).
- Abandonou o open source (biblioteca arquivada) e, em 2026, mudou o foco para segurança de agentes; o produto principal mudou de dono.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ● | RBAC (inclusive global e custom roles por dado), ReBAC (relações, herança), ABAC (atributos como fatos, condicionais), polimorfismo (`extends`) |
| 2. Decisões | ● | authorize, actions (quais ações), list (recursos), query (quem tem acesso, consultas arbitrárias), bulk; Explain na UI |
| 3. Políticas | ◐ | Polar com blocos de recurso; `oso-cloud validate`/`test`; Workbench visual que gera Polar; ambientes; CI documentado. Sem versionamento nativo de política além de PITR (30 dias) e Git do cliente; sem implantação gradual |
| 4. Dados | ● | Fatos centralizados, context facts (até 100 por requisição), Local Authorization (Postgres, MySQL; documentos em beta), Oso Sync (reconciliação a partir de Postgres, MySQL, MongoDB, CSV) |
| 5. Acesso dinâmico | ◐ | Acesso com expiração (`has_role_with_expiry` + `@current_unix_time`). Sem solicitações/aprovações, M de N, delegação ou obrigações nativas |
| 6. Estados e transições | ○ | Estado só como fato/atributo usado em regra; sem transições |
| 7. Padrões | ○ | AuthZEN não encontrado (busca em llms.txt e llms-full.txt da doc). Sem RAR/UMA |
| 8. Máquinas e agentes | ◐ | Atores de qualquer tipo; novo produto "Oso for Agents" (auditoria de sessões, inventário de MCP) é outro produto, não PDP |
| 9. Desempenho e resiliência | ◐ | SaaS multirregião (99,99% declarado); fallback node read-only no cliente (sincroniza a cada 30 min); self-hosted só AWS |
| 10. Auditoria | ◐ | Logs de leitura (decisões) e escrita (fatos) na UI; **exportação de logs não disponível**; webhook de escrita configurado pela Oso, "inadequado para compliance" |
| 11. Experiência e administração | ◐ | UI com Rules Editor/Workbench, Fact Schema, Logs, Explain, Recovery. Sem UI embutível para o cliente final |

## Diferenciais / exclusivos
- **Local Authorization**: a política é compilada em SQL (consulta completa ou fragmento de WHERE) a partir de um YAML que mapeia fatos para tabelas. Resolve a sincronização para dados de alta cardinalidade sem copiar dados.
- **List filtering híbrido** (`list_local`): fatos centralizados (papéis) + relações locais (posse) num só WHERE. Faixa declarada: até ~10 mil recursos (centralizado) vs. milhões (local).
- **Testes na própria linguagem** (`test`, `setup`, `assert`, `assert_not`, `iff`, fixtures).
- **Oso Migrate**: comparação lado a lado com o sistema legado (`ParityHandle`), replay de requisições após mudança de política, depurador em árvore, exportação de requisições logadas como testes.
- **Point-in-time recovery** de política + fatos (30 dias), em novo ambiente ou rollback in-place.

## O que a comunidade elogia (com evidência)
- Artigos técnicos ("Why Authorization Is Hard", 303 pontos no HN, 2021) deram autoridade ao time no tema (HN, comunidade).
- Explain + Logs ligados ("Try this query") encurtam a depuração (doc do fornecedor).
- Não encontrei elogios independentes recentes (busca web e HN; Reddit não acessível por API nesta sessão).

## Do que a comunidade se queixa (com evidência)
- Issues mais votadas da biblioteca legada eram por **mais linguagens** (.NET #596, PHP/Laravel #791, Python async #1295, aarch64 #1529), **custom roles** (#624) — muitas abertas desde 2020–2022 e nunca resolvidas antes do arquivamento.
- **Risco de fornecedor**: a biblioteca OSS foi descontinuada em favor do SaaS e arquivada; o SaaS foi vendido a uma holding e o time saiu (osohq.com, GeekWire, Temporal). Preocupação recorrente no HN sobre depender de autorização remota de startups (Launch HN Warrant, 2022).
- Limitações documentadas do Local Authorization: não suporta lógica que concede acesso a todos os recursos via dado do banco, desigualdades entre dado da nuvem e do banco, recursão misturando fatos locais e centralizados.
- `list` exige variáveis restritas a um domínio de fatos (erro com regras sobre usuários não restritos).

## Histórico de segurança
- OSV: nenhuma vulnerabilidade para `oso` (PyPI, npm, crates.io), `go-oso`, `oso-cloud` (PyPI, npm).
- Nenhum advisory encontrado nas fontes consultadas.
- **Lição:** o risco relevante aqui não é CVE e sim **continuidade**: arquivamento de biblioteca usada em produção sem correções de segurança futuras. Para nós: compromisso público de suporte/LTS e governança que sobreviva à empresa.

## Arquitetura e decisões notáveis
- Oso Cloud replica regras e dados entre regiões; clientes chamam API HTTP.
- Fallback node: contêiner read-only, sincroniza política e fatos a cada 30 min; armazenamento efêmero (falha ao iniciar se a nuvem cair) ou persistente; snapshots exportáveis; SDK troca para o fallback em erro de conexão ou 5xx, **não** em erro de DNS (decisão intencional para não mascarar configuração errada).
- Sessão com afinidade recomendada (nós podem divergir até 1 min).
- Limites (Startup/Growth): leitura sem rate limit; 10 MiB por requisição; 100 context facts; execução até 20 s; escrita 30.000 fatos/min; 1.000 entradas por batch.
- Self-hosted: só AWS, "exige recursos de engenharia do cliente".

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy modelada na política (tenant como recurso pai; papéis por organização). Ambientes separados por conta.
- Bancos da aplicação: Local Authorization com PostgreSQL e MySQL; MongoDB, Pinecone, Elasticsearch em beta. Oso Sync: Postgres, MySQL, MongoDB, CSV.
- Soberania: dados centralizados ficam no Oso Cloud; Local Authorization mantém dados no banco do cliente (Oso "não lê seus dados", só gera SQL).

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Linguagem:** Polar (lógica declarativa, resource blocks com `roles`, `permissions`, `relations`, regras de herança).
- **Consistência:** não documentada com tokens; fatos centralizados com escrita síncrona via API; fallback com defasagem de até 30 min.
- **Listagem:** `list`, `listPaginated`, `list_local` (SQL), `query` reversa (quem tem acesso).
- **Sincronização (foco):** recomendação explícita de **escrita dupla + reconciliação periódica** (`oso-cloud reconcile --perform-updates`, com queries SQL por tipo de fato e modo dry-run), ou evitar sincronizar com Local Authorization / context facts. Guia "adopt local authorization" em etapas.
- **Explicabilidade:** Explain mostra fatos e regras que levaram à decisão.
- **Testes:** testes Polar, Dev Server em CI, Oso Migrate (paridade).
- **Avaliação local:** não há motor embutido no SDK; "local" significa dados locais (SQL), não decisão local. Fallback node é o único PDP no ambiente do cliente.
- **Desempenho publicado:** não encontrado número oficial de latência na doc.
- **AuthZEN:** não encontrado.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N):** **não.** Nenhum recurso de solicitação/aprovação na documentação.
- **Autorização sensível ao estado / transições declarativas:** **não.** Estado pode ser fato/atributo em regra Polar, mas não há modelo de estados nem "pode fazer a transição T no estado S".

## Implicações para o projeto
1. **Adotar** a ideia de "dados onde estão": compilar políticas para filtros SQL (listagem em grande volume sem sincronizar). Forte candidato para a camada de listagem (5.2.2) e para o problema da escrita dupla.
2. **Adotar** testes declarados junto da política (com fixtures de fatos) e ferramenta de paridade/replay para migração de sistemas legados.
3. **Adotar** PITR de política + dados por ambiente (rollback de governança).
4. **Evitar** dependência de serviço remoto sem PDP local de primeira classe: fallback com 30 min de defasagem é paliativo.
5. **Evitar** o caminho "OSS → SaaS-only": premissa 11 exige que a edição aberta seja completa; o arquivamento do Oso é argumento de venda para nós.
6. **Oportunidade**: exportação de logs de decisão e auditoria configurável (o Oso não exporta logs).
7. **Oportunidade**: linguagem lógica é poderosa mas proprietária; preferir linguagem existente/aberta ou contrato AuthZEN na frente.

## Fontes
- https://www.osohq.com/pricing — 2026-11-03 — fornecedor (só "Oso for Agents")
- https://www.osohq.com/post/oso-has-been-acquired-by-tiny — 2026-11-03 — fornecedor (06/10/2026)
- https://temporal.io/blog/oso-security-joins-temporal-accelerate-ai-agent-security — 2026-11-03 — parte interessada (via resultado de busca)
- https://geekwire.com/2026/fast-growing-temporal-hires-team-from-nyc-startup-oso-to-build-ai-agent-security-controls/ — 2026-11-03 — imprensa (via resultado de busca)
- https://github.com/osohq/oso — 2026-11-03 — repositório (arquivado em 22/10/2026)
- https://github.com/osohq/oso/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03 — comunidade
- https://www.osohq.com/docs/develop/facts/local-authorization — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/develop/enforce/list-filtering — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/deploy/fallback-nodes — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/deploy/deployment-models — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/reference/polar/tests — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/develop/facts/sync-facts — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/deploy/ci-cd — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/develop/troubleshooting/debugging.md — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/develop/troubleshooting/logs.md — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/develop/local-dev/oso-migrate.md — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/deploy/backups-and-recovery.md — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/deploy/account-management/rate-limits (via llms-full.txt) — 2026-11-03 — fornecedor
- https://www.osohq.com/docs/llms.txt e /docs/llms-full.txt — 2026-11-03 — fornecedor (busca por "authzen": 0 ocorrências)
- https://www.osohq.com/post/rust-rocket-sqlx — 2026-11-03 — fornecedor (stack)
- https://toolradar.com/tools/oso/pricing — 2026-11-03 — terceiro, baixa confiabilidade (Startup ~US$ 500/mês; não confirmado)
- https://news.ycombinator.com/item?id=28543457 ("Why Authorization Is Hard") — 2026-11-03 — comunidade
- https://news.ycombinator.com/item?id=33081068 (Oso Cloud GA) — 2026-11-03 — comunidade
- https://api.osv.dev/v1/query (oso, go-oso, oso-cloud) — 2026-11-03 — base neutra
