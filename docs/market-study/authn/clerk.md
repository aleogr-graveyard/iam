# Clerk

| Campo | Valor |
|---|---|
| Categoria | IDaaS focado em desenvolvedores (componentes de UI prontos + APIs), com B2B (Organizations), cobrança (Billing) e autenticação de máquinas |
| Site / repositório | https://clerk.com · docs https://clerk.com/docs · SDKs em https://github.com/clerk (serviço fechado; SDKs abertos) |
| Licença e modelo de negócio | Serviço proprietário; SDKs abertos (licença dos SDKs não verificada: GitHub inacessível nesta sessão). Planos Hobby (grátis, 50 mil MRU/app), Pro (US$ 25/mês), Business (US$ 300/mês), Enterprise. Cobra por **MRU** (usuário que volta 24h+ após o cadastro) e **MRO** (organização com ≥2 membros e ≥1 MRU). Add-on **B2B Authentication** (US$ 100/mês): membros ilimitados, papéis customizados, *role sets*, domínios verificados, SSO ligado a organização. SSO e Directory Sync: 1 conexão inclusa no Pro, depois US$ 75/mês cada. MFA, passkeys e senha customizada só em plano pago |
| Stack | Fechada. Postgres (confirmado pelo postmortem de fev/2026); Google Cloud + Cloudflare (página de segurança). Linguagem do backend: não verificada |
| Maturidade | Desde ~2020 (domínio clerk.dev). `@clerk/backend` ~21,7 mi downloads/mês e `@clerk/nextjs` ~10,6 mi (npm, out/2026). SOC 2 Type 2 e HIPAA desde mai/2022 |
| Acesso | 2026-11-03 |

> Limitações: GitHub (issues, estrelas) e Reddit bloqueados nesta sessão; advisories obtidos via OSV.dev (espelho do GitHub Advisory Database).

## Resumo
- Autenticação "drop-in" com componentes React/Next.js (`<SignIn/>`, `<OrganizationSwitcher/>`), forte em DX e em Next.js.
- Modelo de sessão híbrido: sessão com estado no servidor + **token de sessão JWT de 60 s** renovado em segundo plano.
- B2B por **Organizations** sobre um *pool* único de usuários: uma pessoa, várias organizações, papel por organização, "organização ativa" na sessão.
- Não oferece isolamento total por cliente no mesmo app (modelo "Platforms" via apps separados) nem residência de dados (só EUA).
- Histórico recente de CVEs em **checagens de autorização dos SDKs** e de indisponibilidades (postmortem fev/2026).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | Senha (checada contra HIBP), código e link por e-mail, código SMS (pago), passkeys, ~30 provedores sociais, Web3, *sign-in tokens*, SSO corporativo SAML/OIDC/EASIE. LDAP/AD direto e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ◐ | TOTP, SMS, códigos de backup; passkey satisfaz MFA (apps novos desde jul/2026) e pode ser 2º fator na web (out/2026). *Reverification* (step-up) existe. Push, WhatsApp, MFA adaptativo por risco: não encontrados (há *Device Trust* pedindo 2º fator em dispositivo novo) |
| 3 | Ciclo de vida da conta | ● | Cadastro, verificação, recuperação, vinculação automática de contas, convites (individuais e em lote), *waitlist*, allowlist/blocklist, ban, lockout, exclusão, Directory Sync (SCIM e Google Workspace), JIT. Verificação de idade: não encontrada |
| 4 | Sessões e tokens | ◐ | Lista de dispositivos e revogação, sessões simultâneas (multi-conta), duração 5 min–10 anos (pago), JWT templates. DPoP, back-channel logout: não encontrados |
| 5 | Protocolos | ◐ | Clerk como provedor OAuth/OIDC (OAuth applications), device authorization grant, Client ID Metadata Documents; SAML só como SP; SCIM de entrada. CIBA, PAR, FAPI: não encontrados |
| 6 | Máquinas e agentes | ◐ | *Machines* com escopos e M2M tokens; API keys por usuário/organização; OAuth para servidores MCP; servidor MCP próprio para configurar a Clerk. Identidade de workload federada: não encontrada |
| 7 | Detecção e defesa | ● | "Protect": lockout por força bruta, Turnstile, checagens de dispositivo, bloqueio de e-mail descartável; avançado (Business): rate limits, bloqueio por tipo de rede/país. Allowlist de IP para a secret key |
| 8 | Privacidade e conformidade | ◐ | SOC 2 Type 2, HIPAA (BAA no Enterprise), GDPR via EU-US DPF, CCPA. **Sem residência de dados** (EUA). ISO 27001: não. Export completo de dados em todos os planos |
| 9 | Experiência e administração | ● | Componentes prontos e "Elements", Account Portal hospedado, i18n, webhooks, impersonação (5 no Hobby), Application Logs e Admin Logs (Business) |
| 10 | Operação e migração | ◐ | Só SaaS; ambientes dev/prod por instância; SLA 99,99% só no Enterprise; postmortems públicos |
| 11 | Identidade descentralizada | ○ | Não encontrado (Web3 wallets não são VC/EUDI) |

## Diferenciais / exclusivos
- **Token de sessão de 60 s + refresh silencioso**: limita a janela de abuso de token roubado sem perder verificação local (sem ida ao banco).
- **Role Sets**: conjuntos de papéis atribuídos por organização (ex.: por plano de preço ou por coorte de cliente), com fluxo guiado de **remapeamento** quando um papel some do conjunto.
- **Domínios verificados com dois modos**: convite automático ou sugestão com pedido de entrada aprovado por admin (*membership request*).
- **Session tasks**: pendências pós-login (escolher organização, configurar MFA) antes da sessão ficar "ativa".
- **Billing integrado** a usuários e organizações (planos por assento).
- **Clerk for Platforms**: API para provisionar uma aplicação Clerk por cliente da plataforma e *claimable apps* (o cliente assume a conta).

## O que a comunidade elogia (com evidência)
- DX e rapidez de integração em Next.js — comentários no Launch HN do Stack Auth comparando "Clerk seems fine, Auth0 is fine" e o posicionamento "we are to Ory what Clerk is to Auth0" (HN 41194673).
- Postmortems detalhados e públicos (fev/2026).

## Do que a comunidade se queixa (com evidência)
- **Preço para projetos pequenos / lock-in**: "Clerk's pricing got a bit out of hand for side projects" (HN 48047649, mai/2026); "the auth business optimizes for deep lock-in" citando clerk.dev (HN 49480316, ago/2026). O fundador respondeu a acusação de preço "predatório" (HN 40034138, 2024).
- **Disponibilidade**: postmortem de 19/fev/2026 (90 min, >95% das requisições com 429) admite "too many times in recent months" e comunicação de status ruim.
- **Recursos B2B atrás de add-on**: org com no máximo 5 membros por padrão (20 sem add-on), papéis customizados e domínios verificados pagos (docs oficiais).
- Sem residência de dados — limita clientes regulados fora dos EUA (página de segurança).

## Histórico de segurança

| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVE-2024-22206 / GHSA-q6w5-jg5q-47vg | 2024 | Crítica, IDOR/autenticação quebrada em `auth()`/`getAuth()` do `@clerk/nextjs`; a refatoração quebrou a lógica e não havia teste e2e para esse vetor | Testes de regressão de segurança para cada vetor conhecido; refatoração de código de autenticação exige suíte negativa |
| CVE-2025-53548 / GHSA-9mp4-77wg-rwx9 | 2025 | Alta, verificação de assinatura de webhook aceitava eventos mal assinados (CWE-345) | Verificação de webhooks com biblioteca testada, comparação de todas as assinaturas, testes com assinaturas inválidas |
| CVE-2025-63700 / GHSA-3mm3-wfpv-q85g | 2025 | Média, desvio do fluxo OAuth manipulando a etapa de verificação OTP no `clerk-js` | Estado do fluxo deve ser validado no servidor; o cliente não decide transições |
| CVE-2026-34076 / GHSA-gjxx-92w9-8v8f | 2026 | Alta, SSRF no proxy opcional da Frontend API vazava a secret key para host do atacante | Proxies montados com caminho do usuário são superfície de SSRF; nunca encaminhar segredos para host derivado de entrada |
| CVE-2026-41248 / GHSA-vqx2-fgx2-5wq9 | 2026 | Crítica, `createRouteMatcher` contornável com requisições forjadas (normalização de caminho) | Proteção só em middleware por padrão de rota é frágil; checar também no handler; normalização canônica de caminho |
| CVE-2026-42349 / GHSA-w24r-5266-9c3c | 2026 | Alta, `has()`/`auth.protect()` retornavam `true` quando várias condições (organização, billing, reverificação) eram combinadas | Predicados compostos de autorização precisam de testes por combinação (property-based); falhar fechado |
| (Next.js) CVE-2025-29927 | 2025 | Desvio de middleware no Next.js afetando quem dependia só do middleware | Defesa em profundidade além do framework |

Incidente operacional: fev/2026, *query plan flip* do Postgres após `ANALYZE` automático derrubou o serviço; o failover de sessão não disparou porque o banco estava "degradado, não fora". Lição: failover por sintoma (latência/erros), estatísticas estáveis e alertas de mudança de plano.

## Arquitetura e decisões notáveis
- Frontend API (FAPI) dedicada por aplicação + Backend API; cookies `__client` (longo, no domínio da FAPI) e `__session` (60 s).
- Organização ativa é **por aba** do navegador, mas o cookie de sessão é único; a doc alerta para não confiar no cookie em buscas em segundo plano multi-tenant.
- Modo "accountless" e CLI/skills para agentes de código configurarem instâncias.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Modelo B2B**: aplicação (instância) → *pool* único de usuários → Organizations; usuário em N organizações, papel por membership; *Personal Accounts* opcionais; desde ago/2025 "membership required" é o padrão.
- Papéis/permissões definidos **no nível da aplicação** e disponibilizados por organização via Role Sets; papéis de sistema (`org:sys_*`) + permissões customizadas `org:<recurso>:<ação>`.
- Por organização: domínios verificados (até 10), conexão de SSO, limite de membros, metadados, slug, logo. Não há política de autenticação/MFA por organização documentada (MFA é aplicação inteira).
- **Isolamento total** (pool separado por cliente) não é suportado no mesmo app; a doc chama isso de cenário "Platforms" e resolve com uma aplicação por cliente via Platform API.
- Extensão: metadados (public/private/unsafe), JWT templates, webhooks; sem hooks síncronos no fluxo de login encontrados.
- Bancos: não se aplica (SaaS). Soberania: só EUA.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não**. Só aprovação individual de *membership request* e *waitlist* (docs Verified Domains, Restricting access).
- Autorização sensível ao estado / transições: **não**. RBAC simples por organização; os CVEs de 2026 mostram fragilidade até em predicados compostos.

## Implicações para o projeto
1. **Adotar** "organização ativa" explícita no token e na API, mas por requisição (não só cookie global), evitando a armadilha documentada de múltiplas abas.
2. **Adotar** Role Sets (catálogo de papéis global, subconjunto por organização) com remapeamento guiado; casa com o modelo tenant → organização do brief 4.7.
3. **Adotar** domínios verificados com modos convite/sugestão + aprovação; rejeitar domínios públicos/descartáveis.
4. **Oportunidade**: isolamento total + promoção de organização a tenant no mesmo produto (Clerk exige outra aplicação).
5. **Oportunidade**: residência de dados e self-hosting; Clerk é só EUA.
6. **Evitar** gating de MFA/passkeys/papéis por plano: o brief exige segurança por padrão.
7. **Evitar** proteção baseada só em matcher de rota; testar combinações de predicados (CVE-2026-41248/42349).
8. **Adotar** tokens de sessão muito curtos com refresh, e failover de emissão de token independente do banco principal.

## Fontes
- https://clerk.com/llms.txt e https://clerk.com/docs/llms.txt — 2026-11-03 — oficial
- https://clerk.com/docs/guides/organizations/overview.md, /configure.md, /control-access/roles-and-permissions.md, /control-access/role-sets.md, /add-members/verified-domains.md — 2026-11-03 — oficial
- https://clerk.com/docs/guides/how-clerk-works/overview.md e /multi-tenant-architecture.md — 2026-11-03 — oficial
- https://clerk.com/docs/guides/configure/auth-strategies/sign-up-sign-in-options.md — 2026-11-03 — oficial
- https://clerk.com/pricing.md — 2026-11-03 — oficial
- https://clerk.com/security.md — 2026-11-03 — oficial
- https://clerk.com/platform — 2026-11-03 — oficial (marketing)
- https://clerk.com/blog/2026-02-19-system-outage-postmortem.md — 2026-11-03 — oficial
- https://api.osv.dev/v1/vulns/GHSA-q6w5-jg5q-47vg, GHSA-9mp4-77wg-rwx9, GHSA-3mm3-wfpv-q85g, GHSA-gjxx-92w9-8v8f, GHSA-vqx2-fgx2-5wq9, GHSA-w24r-5266-9c3c — 2026-11-03 — base pública (GitHub Advisory)
- https://clerk.com/blog/cve-2025-29927 — 2026-11-03 — oficial (via busca)
- https://api.npmjs.org/downloads/point/last-month/@clerk/backend — 2026-11-03 — npm
- https://news.ycombinator.com/item?id=41194673, 48047649, 49480316, 40034138, 47100047 — 2026-11-03 — comunidade (41194673 é lançamento de concorrente)
