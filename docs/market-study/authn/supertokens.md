# SuperTokens

| Campo | Valor |
|---|---|
| Categoria | Framework de autenticação "híbrido": serviço núcleo (SuperTokens Core) + SDKs de backend e frontend embutidos na aplicação; self-hosted ou gerenciado |
| Site / repositório | https://supertokens.com · https://github.com/supertokens/supertokens-core · SDKs `supertokens-node`, `supertokens-python`, `supertokens-golang`, frontend `supertokens-auth-react`/`web-js` |
| Licença e modelo de negócio | Open core. Núcleo Apache-2.0, mas o diretório `ee/` está sob "SuperTokens Enterprise License" (uso em produção só com assinatura; modificações pertencem à SuperTokens). Recursos pagos são ligados por chave de licença via feature flags: `account_linking`, `multi_tenancy`, `dashboard_login`, `mfa`, `security`, `oauth`, `saml` (arquivo `EE_FEATURES.java`). Cobrança por MAU (ex.: MFA US$ 0,01/MAU, mínimo US$ 100/mês — valores de terceiros, não confirmados no site) |
| Stack | Core em Java (JDK embutido, Tomcat embutido); SDKs Node, Python, Go. Banco: PostgreSQL (changelogs recentes só citam PostgreSQL ≥ 14), plugin MySQL ainda publicado (migrações citadas até a 10.1.x), armazenamento em memória (SQLite) para dev; MongoDB só em versões antigas. OAuth2 provider apoiado em **Ory Hydra** (serviço separado, portas 4444/4445) |
| Maturidade | Empresa YC S20 (Launch HN 2020); core 12.4.1; `supertokens-node` 24.1.0 (29/10/2026). ~15,3 mil estrelas, 154 issues abertas, OpenSSF Scorecard 5/10 (deps.dev, 24/08/2026). ~75 mil downloads/semana do SDK Node (npm) |
| Acesso | 2026-11-03 |

> Limitação: `supertokens.com` não estava liberado na rede da sessão. Usei o repositório (README, CHANGELOG, `config.yaml`, `ee/LICENSE.md`), advisories do GitHub, npm, deps.dev e buscas (várias fontes de terceiros e espelhos de docs).

## Resumo
- Arquitetura em três peças: frontend SDK (UI pronta e gestão de tokens), backend SDK (expõe as APIs na sua aplicação, verifica sessão localmente) e Core (Java, HTTP, lógica e banco).
- Ponto forte histórico: **gestão de sessão** (access token curto + refresh token rotativo com detecção de roubo, cookies HttpOnly, anti-CSRF).
- "Recipes" cobrem senha, passwordless, social, passkeys, MFA, multi-tenancy, papéis, SAML e OAuth2 provider — mas os recursos de empresa são pagos mesmo self-hosted.
- Outubro/2026: **15 advisories** publicados de uma vez (8 no core, 7 no SDK Node), vários altos/críticos, envolvendo SAML, isolamento entre tenants e vinculação de contas.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | E-mail/senha, telefone/senha, passwordless (link mágico e OTP por e-mail/SMS), social, passkeys (recipe WebAuthn), SAML (pago, `saml`), OIDC corporativo via third-party por tenant. LDAP/AD e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ◐ | MFA pago: TOTP, OTP por e-mail/SMS, fatores por tenant (`first_factors`). Push e MFA adaptativo: não encontrados. Códigos de recuperação: não verificado |
| 3 | Ciclo de vida da conta | ◐ | Cadastro, verificação de e-mail, reset, vinculação de contas (paga), metadados, papéis, importação em massa (bulk import), exclusão. Bloquear/banir usuário: issue #521 aberta. SCIM, convites, verificação de idade: não encontrados |
| 4 | Sessões e tokens | ◐ | Rotação de refresh token com detecção de reuso (revoga a família), anti-CSRF, sessões listáveis/revogáveis, JWT e JWKS. DPoP, back-channel logout: não encontrados |
| 5 | Protocolos | ◐ | OAuth2/OIDC provider (pago, `oauth`, via Ory Hydra): authorization code, refresh, client credentials, introspecção, revogação; SAML (IdP e SP, pago); WebAuthn. Device flow, CIBA, PAR, FAPI, SCIM: não encontrados |
| 6 | Máquinas e agentes | ◐ | M2M via client credentials (tabela `oauth_m2m_tokens` no changelog). Chaves de API para usuários finais, identidade de workloads, agentes de IA: não encontrados |
| 7 | Detecção e defesa | ◐ | Flag paga `security` (suíte de proteção; conteúdo não verificado). Rate limiting em APIs: issue #163 aberta. Senhas vazadas/bots: não verificado |
| 8 | Privacidade e conformidade | ? | Self-hosting dá soberania; `activity_log` com retenção configurável (31 dias). Perfis LGPD/GDPR: não encontrados |
| 9 | Experiência e administração | ◐ | UI pré-construída (React) ou headless; User Management Dashboard (usuários além de 3 pagos); multi-idioma da UI: não verificado. Webhooks: não encontrados |
| 10 | Operação e migração | ◐ | Bulk import com hashes, scripts de migração SQL por versão, porta admin separada, `/livez`. Backup = banco do operador |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Sessão como produto**: rotação de refresh token com detecção de roubo por reuso da família, defendida pelo fundador como alternativa a fingerprint/IP ([blog](https://supertokens.com/blog/session-hijacking-fixation), fornecedor).
- **Verificação de sessão sem chamar o core**: o backend SDK valida o access token localmente; o core só é chamado em refresh/criação (README).
- **Overrides de API e de função** nos SDKs (sobrescrever qualquer endpoint ou função de recipe) — extensibilidade por código sem fork.
- **Multi-tenancy em camadas**: connection URI domain → app → tenant, com storage opcional por tenant e cotas de concorrência por domínio (`max_concurrent_requests_per_cud`, 12.4.0).
- **`activity_log` transacional**: eventos de ciclo de vida gravados atomicamente com a mutação, garantido por um guard AspectJ em tempo de compilação (12.x).

## O que a comunidade elogia (com evidência)
- Self-hosted gratuito sem limite de usuários para os recursos abertos; boa alternativa a Auth0/Cognito/Firebase ([HN 35895840](https://news.ycombinator.com/item?id=35895840), 212 pts).
- Abertura de partes do código permite fork ("you can go right ahead and fork the open-source parts", idem, 35901127).
- Changelog extremamente detalhado e foco em desempenho/concorrência (pools de conexão, testes de 1 M usuários) — [CHANGELOG](https://github.com/supertokens/supertokens-core/blob/master/CHANGELOG.md).

## Do que a comunidade se queixa (com evidência)
- **Licença/open core**: a licença `ee/` proíbe copiar/distribuir; "key features ... are not open source" ([HN 35895840](https://news.ycombinator.com/item?id=35895840), comentários 35898613 e 35898693).
- **Java**: "can't afford hiring a jvm specialist" (idem, 35898300); "Stopped reading at 'Why Java'" (35903524).
- **Sem OIDC à época** = aprisionamento (35900074) — depois resolvido com o OAuth2 provider (pago, sobre Hydra).
- **Risco de empresa pequena** (35903816).
- **Estabilidade e operação self-hosted**: "I keep regularly hitting tons of bugs", erros "No SuperTokens core available", incompatibilidade entre versões do core ([fórum da comunidade](https://community.supertokens.com/t/2150542/how-stable-is-supertokens-i-keep-regularly-hitting-tons-of-b), 2021–2024).
- **Issues mais votadas** ([GitHub](https://github.com/supertokens/supertokens-core/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc)): SDK C#/.NET (#401), Elixir (#13), Deno (#399), Spring Boot (#420), banir usuário (#521), rate limiting (#163), métricas (#244), OU/grupos em papéis (#468), multi-tenancy com pool de usuários por tenant sem infraestrutura redundante (#283), informações de IP/dispositivo na sessão (#55).

## Histórico de segurança
Fontes: [advisories do core](https://github.com/supertokens/supertokens-core/security/advisories) (8, todos de 16/10/2026) e [do SDK Node](https://github.com/supertokens/supertokens-node/security/advisories) (7, 22/10/2026). OSV não lista nenhum (sem CVE/ecossistema mapeado na data).

| ID | Ano | Severidade / classe | Lição para nós |
|---|---|---|---|
| GHSA-4rf2-xqvf-8rcx | 2026 | Crítica — API key de app sobrescreve credenciais do domínio pai e toma apps irmãos | Hierarquia de tenants exige autorização por nível; chave de escopo menor nunca altera escopo maior |
| GHSA-q7mc-f2jm-r4gr | 2026 | Alta — caminho de tenant desconhecido burla API key | Tenant inexistente = negar (falhar fechado) |
| GHSA-6v34-w48w-44rc | 2026 | Alta — passkey aceita entre tenants | Credenciais sempre com escopo de tenant/RP na consulta |
| GHSA-gvx2-mwmj-wwvj | 2026 | Alta — XML Signature Wrapping no SAML | Confirma a exceção da 4.30: SAML só com biblioteca consolidada e testes de XSW |
| GHSA-wfg2-vqfg-rqh6 | 2026 | Alta — identidade lida de asserção que não passou na checagem de audiência | Extrair identidade só da asserção validada (uma única) |
| GHSA-3r7x-rhvc-fw28 | 2026 | Alta — replay de RelayState/SAMLResponse | Uso único de asserções (`InResponseTo`, tabela de consumidas), pinagem de Destination/Recipient |
| GHSA-w9fp-wv6g-pqv8 | 2026 | Alta — sessão do dashboard sem checagem de expiração | Sessões administrativas com o mesmo rigor (ou maior) que as de usuários |
| GHSA-2pc4-vw9j-454c | 2026 | Moderada — sessão criada em paralelo ao unlink sobrevive | Revogação e criação serializadas por usuário |
| GHSA-93m2-695m-88qg (node) | 2026 | Alta — vinculação automática sem checar e-mail verificado do usuário primário | Boa prática do brief 5.1 (vincular só com e-mail verificado) |
| GHSA-7xgh-95v6-9224 (node) | 2026 | Alta — recuperação via WebAuthn usa opções geradas para outro e-mail | Vincular desafio WebAuthn ao usuário/fluxo |
| GHSA-j9qp-rxcc-p53p (node) | 2026 | Alta — callback `form_post` da Apple envia code/id_token a URL do atacante | Redirect allowlist também em `form_post` |
| GHSA-496x-gccr-82xf (node) | 2026 | Alta — corpo comprimido sem limite (bomba de descompressão) | Limites de tamanho antes de descomprimir |
| GHSA-cwq8-r97g-9cvh (node) | 2026 | Alta — query string burla restrição de escrita do dashboard | Autorização sobre rota normalizada |

Lição geral: concentração de falhas em **SAML** e **fronteiras de tenant** — exatamente as áreas de maior risco do nosso projeto (premissa 10). A publicação simultânea de 15 advisories sugere auditoria/pentest recente (não confirmado).

## Arquitetura e decisões notáveis
- Core stateless em Java + banco; vários cores atrás de balanceador; SDKs falam com o core por HTTP com API key.
- Versionamento de contrato entre SDK e core via **CDI** (Core Driver Interface), com matriz de compatibilidade.
- OAuth2/OIDC provider delega a Ory Hydra (serviço extra para operar) em vez de implementar o protocolo.
- Ênfase recente em robustez: detecção de aquisição aninhada de conexões, `READ COMMITTED` + `SELECT ... FOR UPDATE`, porta admin isolada para health check.
- Migrações SQL aplicadas no boot, com scripts offline para bases grandes.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy (paga): por tenant, métodos de login, provedores, fatores de MFA e, opcionalmente, banco próprio; usuários podem pertencer a vários tenants.
- Extensibilidade: overrides de APIs/funções nos SDKs; plugin de storage em Java.
- Configuração: `config.yaml`/variáveis de ambiente no core + configuração por app/tenant via API (core configs por tenant) — mais próximo da premissa 3 do que bibliotecas puras.
- Bancos: PostgreSQL (foco), MySQL (plugin), memória para dev.
- Soberania: total no self-hosted; licença paga verificada por sincronização periódica com a SuperTokens (cron `EELicenseCheck`) — dependência externa a considerar.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** (nada encontrado em recipes, dashboard ou changelog).
- **Autorização sensível ao estado / transições declarativas**: **não**. Autorização limitada a recipe UserRoles (papéis e permissões, RBAC) — issue #204 "Access control recipe".

## Implicações para o projeto
1. **Adotar** rotação de refresh token com detecção de reuso por família como padrão (RFC 9700/OAuth 2.1), incluindo testes de concorrência.
2. **Adotar** verificação local de token nos SDKs (sem ida ao serviço) com contrato versionado entre SDK e servidor (equivalente ao CDI).
3. **Adotar** auditoria de ciclo de vida gravada na mesma transação da mutação (outbox/ledger) e verificada por ferramenta (lint/teste), como o guard AspectJ.
4. **Evitar** open core que tranca MFA, multi-tenancy e SAML: é a queixa central; nossa premissa 11 deve monetizar operação/serviço, não segurança básica.
5. **Evitar** dependência de serviço externo para OAuth (Hydra) e de checagem de licença online no self-hosted.
6. **Prioridade de testes**: SAML (XSW, replay, audiência) e isolamento entre tenants/apps devem ter suíte ofensiva própria desde o primeiro marco que os incluir.

## Fontes
- https://github.com/supertokens/supertokens-core (README) — 2026-11-03
- https://raw.githubusercontent.com/supertokens/supertokens-core/master/CHANGELOG.md — 2026-11-03
- https://raw.githubusercontent.com/supertokens/supertokens-core/master/config.yaml — 2026-11-03
- https://raw.githubusercontent.com/supertokens/supertokens-core/master/ee/LICENSE.md — 2026-11-03
- https://raw.githubusercontent.com/supertokens/supertokens-core/master/src/main/java/io/supertokens/featureflag/EE_FEATURES.java — 2026-11-03
- https://github.com/supertokens/supertokens-core/security/advisories — 2026-11-03
- https://github.com/supertokens/supertokens-node/security/advisories — 2026-11-03
- https://github.com/supertokens/supertokens-core/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://github.com/supertokens/supertokens-mysql-plugin — 2026-11-03
- https://cdn.jsdelivr.net/npm/supertokens-node@24.0.2/compose.yml (via busca: core apontando para Hydra 4444/4445) — 2026-11-03
- https://mintlify.wiki/supertokens/supertokens-core/auth/oauth (espelho de docs, via busca) — 2026-11-03 — terceiro
- https://supertokens.com/pricing (trechos via busca) — 2026-11-03 — fornecedor
- https://skycloak.io/solutions/vs-supertokens/ — 2026-11-03 — **concorrente**
- https://toolquestor.com/vs/supertokens-vs-better-auth — 2026-11-03 — agregador
- https://supertokens.com/blog/session-hijacking-fixation (via busca) — 2026-11-03 — fornecedor
- https://community.supertokens.com/t/2150542/how-stable-is-supertokens-i-keep-regularly-hitting-tons-of-b — 2026-11-03 — fórum do fornecedor
- https://api.deps.dev/v3/projects/github.com%2Fsupertokens%2Fsupertokens-core — 2026-11-03
- https://registry.npmjs.org/supertokens-node, https://api.npmjs.org/downloads/point/last-week/supertokens-node — 2026-11-03
- https://news.ycombinator.com/item?id=35895840, https://news.ycombinator.com/item?id=24306572 — 2026-11-03
