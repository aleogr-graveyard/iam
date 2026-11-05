# PlainID

| Campo | Valor |
|---|---|
| Categoria | Plataforma corporativa de autorização (PBAC), "Authorization Management Platform": PAP central em SaaS + PDP/PIP híbridos + conectores de enforcement ("Authorizers") |
| Site / repositório | https://www.plainid.com · docs https://docs.plainid.io (índice em https://docs.plainid.io/llms.txt). Código fechado; não há repositório público do produto |
| Licença e modelo de negócio | Proprietário, assinatura corporativa (preço sob consulta, "Get a demo"). Nada aberto além de exemplos e integrações. Imagens de container em Docker Hub privado, assinadas com Cosign |
| Stack | Não publicada. PAP em SaaS multi-região (`{cliente}.{região}.plainid.io`); componentes do cliente (PAA) em Kubernetes (Helm) ou binário standalone; Redis para cache local (Redis Cluster mode não suportado); políticas exportáveis como "Structured Rego" |
| Maturidade | Empresa israelense-americana (fundada ~2014, não verificado nesta sessão). Site declara "mais de 35 milhões de identidades e bilhões de decisões por dia". Citada como Example Technology em "Authorization Management Platform" no Reference Architecture Brief "IAM for AI agents" do Gartner (abr/2026). Gartner Peer Insights: 4,5 com só 5 avaliações |
| Acesso | 2026-11-03 |

## Resumo
- PDP central com **políticas por identidade, ativo, ação e condição**, montadas em UI (Policy Canvas/Policy Builder) a partir de "blocos" (Dynamic Groups, Rulesets, Conditions, Actions).
- Ênfase em **orquestração**: traduz e empurra políticas para sistemas nativos (Snowflake, Databricks, BigQuery, Power BI, Zscaler…) e compara a política central com a do fornecedor ("Vendor Compare").
- Implantação **híbrida**: o PAA (Policy Authorization Agent = PDP + PIP + agente + Redis) roda na rede do cliente; o controle (PAP) fica no SaaS.
- Suporte nativo ao **AuthZEN Authorization API 1.0** e foco recente em agentes de IA (MCP Gateway, LangChain/LangGraph, guardrails, RAG).
- Ciclo de vida de política (versionamento, revisão, promoção) é **delegado ao Git/CI do cliente**; o simulador de políticas é beta e fora do SLA.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | PBAC/ABAC nativos; RBAC como caso particular (Dynamic Groups). ReBAC nativo não encontrado ("Access Graph" é visualização). "Multi-Identity Policy Evaluation" na API |
| 2. Decisões | ● | Permit/Deny (incl. V5 para APIs), Policy Resolution, **User List** (quem acessa), **User Access Token/Asset Resolution** (o que o usuário acessa), Policy List. Explicabilidade via `includeDenyReason` (códigos PID000x) |
| 3. Políticas | ◐ | Autoria em UI; import/export/validate em **Structured Rego**; versionamento e aprovação **fora** do produto (Git/PR); simulador beta; certificação periódica via APIs. Implantação gradual nativa não encontrada |
| 4. Dados | ● | PIP com JDBC, LDAP, REST, Entra ID, Cosmos DB, MongoDB, Snowflake, BigQuery; "Views" com cache; invalidação de cache por API. Consistência por TTL (sem token de consistência) |
| 5. Acesso dinâmico | ◐ | "Zero Standing Privileges" é posicionamento; condições de tempo/IP. Solicitação/aprovação de acesso, JIT com aprovação, delegação e M de N **não encontrados** na documentação. Respostas com filtros de linha/coluna funcionam como obrigações implícitas |
| 6. Estados e transições | ○ | Estado só como atributo de ativo/condição |
| 7. Padrões | ◐ | AuthZEN 1.0 nativo (página oficial); membro do WG. RAR, UMA: não encontrados. Respostas JWT assinadas pelo PDP |
| 8. Máquinas e agentes | ● | Agentic Identity Control: registro e descoberta de agentes, MCP Gateway, guardrails de entrada/saída, RAG, PlainID Edge (no ambiente do cliente) |
| 9. Desempenho e resiliência | ◐ | PDP local (PAA) + caches (identidade, decisão por escopo, JWKS) em Redis; sidecar Envoy/Istio. Números de latência publicados: só "milissegundos" (marketing) |
| 10. Auditoria | ● | Relatórios de auditoria administrativa e de autorização; API de eventos administrativos; configuração de auditoria de autorização no PAA |
| 11. Experiência e administração | ● | UI extensa (workspaces de Identidades, Políticas, Integrações), Policy Canvas, servidor MCP para administração, Management APIs; enriquecimento de token em IdPs (Okta, Ping, Auth0, Entra) |

## Diferenciais / exclusivos
- **Orquestração de políticas para sistemas nativos** (Learn Mode descobre e traduz políticas do fornecedor; Manage Mode publica) e **Vendor Compare** com diff assistido por IA — raríssimo no mercado.
- **"Authorizers" prontos** para gateways (Apigee, Kong, AWS API GW, Azure APIM), malha (Envoy/Istio), dados (Snowflake, Databricks, BigQuery, Trino, Denodo, Power BI) e IdPs (token enrichment).
- **Hierarquia Tenant → Environment → Workspace** com separação de funções por workspace (identidades vs políticas vs integrações).
- **Deny reasons** padronizados com distinção "ativo inexistente" vs "negado por política" (com custo de consulta extra ao PIP, documentado).
- Listagem nos dois sentidos (User List e Asset Resolution) em produto PBAC, algo que motores ABAC puros costumam não oferecer.

## O que a comunidade elogia (com evidência)
- Funcionalidades, desempenho, suporte e integrações (Gartner Peer Insights, 5 avaliações; uma de mai/2026 "strong and complete authorization solution"). Amostra pequena e não verificável.
- Não há comunidade pública (sem GitHub, fórum aberto, HN/Reddit relevantes encontrados).

## Do que a comunidade se queixa (com evidência)
- Não encontrado: Gartner/Capterra sem "dislikes" visíveis; Capterra com dados inconsistentes. Procurado em Gartner Peer Insights, Capterra, buscas web.
- Sinais na própria documentação (inferência, não queixa): simulador "Beta… não coberto pelo SLA"; versionamento/aprovação de políticas fora do produto; `includeDenyReason` pode alterar o fluxo de avaliação; Redis Cluster mode não suportado.

## Histórico de segurança
- NVD (keywordSearch "PlainID"): 0 CVEs. Nenhum advisory público encontrado.
- **Lição:** ausência de CVE em produto fechado não é evidência de segurança; para nós, programa de divulgação + advisories públicos é parte da premissa 5 (auditável). Copiar a prática deles de **imagens assinadas (Cosign)** e documentação de verificação.

## Arquitetura e decisões notáveis
- Separação PAP (SaaS) × PDP/PIP (PAA no cliente): decisões e dados sensíveis ficam locais; o PAA recebe políticas por push do PAP.
- Cache de decisão chaveado pela **requisição inteira** (Scope Cache), com TTL por escopo; condições de IP/tempo **não** entram na chave (documentado) — risco de decisão cacheada errada se usado com contexto dinâmico.
- Import com duas estratégias de reconciliação de blocos (por conteúdo × por nome), mostrando o custo de políticas compostas por blocos compartilhados.
- Respostas do PDP podem ser **JWT assinados** (chave em gestor de segredos).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Tenant = cliente; Environments (dev/test/prod) e Workspaces dentro do tenant. Multi-tenancy do **cliente do cliente** (plataforma B2B2C): não encontrado como conceito explícito.
- Extensão por conectores (PIP) e Authorizers; políticas como código em Rego estruturado.
- Bancos: PIP lê de vários; armazenamento interno não publicado. Redis no PAA.
- Soberania: PAA e PlainID Edge no ambiente do cliente; segredos em Vault, AWS SM, Azure KV, CyberArk CCP.

## (Somente AuthZ) Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Modelos:** PBAC/ABAC, RBAC por grupos dinâmicos; ReBAC não nativo.
- **Linguagem:** UI + "Structured Rego" (exportável/importável; subset de Rego gerado pela plataforma).
- **Consistência:** TTL de cache e invalidação explícita por API; sem tokens.
- **Listagem:** User List e Asset Resolution; filtros de linha/coluna para dados.
- **Sincronização:** consulta sob demanda ao PIP (sem replicar dados), com cache.
- **Explicabilidade:** deny reason codes; Policy Resolution mostra políticas aplicáveis.
- **Testes:** simulador (beta); testes automatizados de política não encontrados.
- **Avaliação local:** PAA no cliente; não há SDK com avaliação em processo.
- **Desempenho publicado:** apenas alegações de marketing.
- **AuthZEN:** 1.0 nativo no PDP.

## Verificação de diferenciais candidatos do projeto
- **Aprovação M de N como recurso de primeira classe?** **Não.** Nenhuma página de aprovação/solicitação de acesso no índice da documentação (llms.txt, 379 entradas). Revisão de políticas é explicitamente delegada a "GitHub pull requests, dedicated workflow tools" (Policy Lifecycle).
- **Autorização sensível ao estado / transições declarativas?** **Não.** Estado só como atributo/condição; sem conceito de transição.

## Implicações para o projeto
1. **Adotar:** listagem nos dois sentidos e deny reasons padronizados (com códigos estáveis) desde o início; são o "mínimo esperado" corporativo.
2. **Adotar:** separação PAP central × PDP local com push de políticas; encaixa no monólito modular + avaliação local (seção 4.29).
3. **Evitar:** cache de decisão cuja chave ignora contexto (IP/tempo). Nossa chave de cache deve incluir todo atributo que a política lê, ou o cache deve ser desligado para políticas contextuais.
4. **Oportunidade:** versionamento, revisão (inclusive M de N para mudança de política) e implantação gradual **dentro** do produto, onde a PlainID delega ao Git do cliente.
5. **Oportunidade:** orquestração para sistemas nativos é diferencial caro; fica fora das primeiras camadas, mas o contrato de exportação de políticas deve permitir isso depois.
6. **Adotar:** respostas de decisão assinadas (JWT) como opção para PEPs que precisam de evidência.

## Fontes
- https://www.plainid.com/ — 2026-11-03 — site do fornecedor (marketing)
- https://docs.plainid.io/llms.txt — 2026-11-03 — índice oficial da documentação
- https://docs.plainid.io/docs/about-the-platform.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/policy-lifecycle.md — 2026-11-03 — oficial (atualizado 2026-08-10)
- https://docs.plainid.io/docs/certification-process-steps.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/about-the-policy-simulator.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/about-paas-and-pips.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/pdp-caching-options.md — 2026-11-03 — oficial
- https://docs.plainid.io/apidocs/working-with-deny-reason.md — 2026-11-03 — oficial
- https://docs.plainid.io/apidocs/policy-evaluation-logic.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/vendor-compare.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/about-tenants.md, https://docs.plainid.io/docs/about-environments.md, https://docs.plainid.io/docs/faq.md — 2026-11-03 — oficial
- https://docs.plainid.io/docs/plainid-edge.md — 2026-11-03 — oficial
- https://lp.plainid.com/standards/authzen — 2026-11-03 — fornecedor
- https://www.plainid.com/newsroom/plainid-named-in-the-2026-gartner-reference-architecture-brief-iam-for-ai-agents-and-other-workloads-report/ — 2026-11-03 — press release do fornecedor
- https://www.gartner.com/reviews/product/plainid-authorization-platform — 2026-11-03 — avaliações de usuários (amostra de 5)
- https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=PlainID — 2026-11-03 — NVD (0 resultados)
