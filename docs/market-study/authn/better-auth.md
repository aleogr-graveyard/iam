# Better Auth (incorpora o Auth.js / NextAuth.js)

| Campo | Valor |
|---|---|
| Categoria | Biblioteca/framework de autenticação embutido na aplicação (TypeScript), não é serviço; núcleo + plugins |
| Site / repositório | https://better-auth.com · https://github.com/better-auth/better-auth · Auth.js: https://github.com/nextauthjs/next-auth |
| Licença e modelo de negócio | Biblioteca MIT (deps.dev). Pago: "Better Auth Infrastructure" (SaaS complementar: dashboard, audit logs com retenção, detecção de bots/credential stuffing/viagem impossível, log drain para SIEM; planos Starter/Pro/Business/Enterprise, valores divergentes entre páginas). Empresa YC (seed US$ 5 mi, jun/2025); **adquirida pela Vercel em 07/07/2026**, que promete manter MIT e agnóstica de framework. Auth.js passou a ser mantido pela equipe Better Auth em 22/09/2025: apenas correções de segurança/urgentes, sem recursos novos |
| Stack | TypeScript (Node, Bun, Deno, edge). Banco da própria aplicação via adaptadores: Kysely embutido (PostgreSQL, MySQL, SQLite/D1, MSSQL), Drizzle, Prisma, MongoDB; "secondary storage" (ex.: Redis) para sessões/rate limit. Hash de senha padrão `scrypt` (nativo do Node), Argon2 configurável |
| Maturidade | npm desde 04/2024; v1.0 em 11/2024; última estável 1.7.15 (01/11/2026), 1.8 em beta. ~30,5 mil estrelas, ~900 issues abertas (deps.dev). 17,5 mi downloads/semana (npm, 26/10–01/11/2026); Auth.js `next-auth` ainda 9,1 mi/semana |
| Acesso | 2026-11-03 |

> Limitação: `better-auth.com` não estava liberado na rede da sessão. A documentação foi lida no próprio repositório (`docs/content/docs/*.mdx` em raw.githubusercontent.com); preços e Infrastructure via buscas na web.

## Resumo
- "Torne fazer sua própria auth tão fácil que você nunca precise de um serviço": a auth roda dentro do backend da aplicação, sobre o banco dela.
- Núcleo pequeno (e-mail/senha, social, sessões) + dezenas de plugins oficiais (2FA, passkey, SSO/SAML, SCIM, organização, OAuth provider, API key, MCP, agent-auth, device authorization etc.).
- Crescimento explosivo (de zero a ~17 mi downloads/semana em 2,5 anos), absorveu o Auth.js e foi comprado pela Vercel.
- O preço da velocidade: **26 advisories em 22 meses**, vários de tomada de conta e concentrados nos plugins de protocolo (OAuth/OIDC provider, organização, magic link).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha, username, magic link, OTP por e-mail, telefone, passkeys (plugin), login social (dezenas de provedores + `generic-oauth`), One Tap, anônimo, SIWE; SSO corporativo OIDC e SAML 2.0 (plugin `sso`). LDAP/AD: não encontrado na lista de plugins oficiais. ICP-Brasil: ○ |
| 2 | MFA e step-up | ◐ | Plugin `2fa`: TOTP, OTP por e-mail/SMS, códigos de recuperação, dispositivos confiáveis; "fresh session" (`freshAge`) para ações sensíveis. Push e MFA adaptativo por risco: não encontrados (detecção de risco só no Infrastructure pago). Passkey é método primário, não fator configurável de 2FA |
| 3 | Ciclo de vida da conta | ◐ | Cadastro, verificação de e-mail, reset, vinculação de contas, convites (plugin `organization`), SCIM (plugin), ban/impersonação (plugin `admin`), exclusão. Verificação de idade: ○. Contas de serviço: issue #3224 aberta |
| 4 | Sessões e tokens | ◐ | Sessões em banco (cookie opaco), listar/revogar, multi-sessão, cache em cookie opcional, JWT/JWKS (plugin), refresh tokens com rotação no `oauth-provider`. DPoP e back-channel logout: não encontrados |
| 5 | Protocolos | ◐ | OAuth 2.1/OIDC provider (PKCE S256, DCR, introspecção, revogação, RP-initiated logout, consentimento), SAML SP, SCIM, WebAuthn, device flow (RFC 8628), CIBA citado no plugin `agent-auth`, CIMD. PAR, FAPI: não encontrados |
| 6 | Máquinas e agentes | ◐ | Plugin `api-key` (inclusive chaves de organização), client credentials no OAuth provider, plugin `mcp`, plugin `agent-auth` (Agent Auth Protocol: descoberta, registro de agentes, capacidades, JWT curtos; marcado instável). Identidade de workloads: ○ |
| 7 | Detecção e defesa | ◐ | Rate limiter embutido (desligado em dev; padrão 100 req/10 s; memória/banco/secondary storage), plugin `captcha`, plugin `have-i-been-pwned`. Bots, credential stuffing, viagem impossível: só no Infrastructure pago |
| 8 | Privacidade e conformidade | ◐ | Dados ficam no banco do cliente (soberania total). Exclusão de usuário existe; perfis de conformidade, registro de consentimento LGPD/GDPR: não encontrados |
| 9 | Experiência e administração | ◐ | Headless (sem tela hospedada), cliente para React/Vue/Svelte etc., plugin `i18n`, plugin `admin` (API), OpenAPI. Dashboard e audit logs: pagos (Infrastructure). Webhooks: não encontrados no núcleo |
| 10 | Operação e migração | ◐ | CLI `generate`/`migrate` de esquema; validação de esquema no boot; guias de migração (Auth.js, Clerk, Supabase etc. — citados em discussões). Backup = banco da aplicação |
| 11 | Identidade descentralizada | ○ | Só SIWE (Ethereum), não VCs/EUDI |

## Diferenciais / exclusivos
- **Embutido no banco do cliente**: zero serviço extra; tabelas da auth convivem com as da aplicação (relações nativas com `user`).
- **Ecossistema de plugins com cliente tipado**: o plugin adiciona endpoints no servidor e métodos tipados no cliente automaticamente (inferência TypeScript ponta a ponta).
- **Agent Auth Protocol** e plugin MCP: aposta cedo em identidade de agentes (adaptadores OpenAPI e MCP, aprovação via device flow/CIBA).
- **Plugins de cobrança** (Stripe, Polar, Autumn etc.) e `last-login-method` — foco em DX de SaaS.
- `resolveUser` no SSO: mapeamento explícito de identidade federada para usuário existente sem depender de casamento por e-mail (resposta direta a ATO por auto-link).

## O que a comunidade elogia (com evidência)
- DX e consistência entre frameworks: "made auth such a non-issue for me" ([HN 45389293](https://news.ycombinator.com/item?id=45389293), 179 pts).
- Dados no próprio banco, sem aprisionamento a SaaS; adaptadores tornam troca "relativamente fácil" ([HN 48819512](https://news.ycombinator.com/item?id=48819512), comentário 48819639).
- Migrações bem-sucedidas a partir de Clerk/Supabase ([HN 48038827](https://news.ycombinator.com/item?id=48038827), 304 pts, 240 comentários).
- Launch HN bem recebido (259 pts); time responde rápido a falhas (patch de CVE-2025-61928 em ~24 h segundo a ZeroPath).

## Do que a comunidade se queixa (com evidência)
- **Governança/aquisição**: "Vercel is not a great steward of open source"; "Open source isn't really open any more. It's just pre-acquisition" ([HN 48819512](https://news.ycombinator.com/item?id=48819512)). Projeto livre absorvido por empresa com VC ([HN 45389293](https://news.ycombinator.com/item?id=45389293)).
- **Fim do Auth.js** sem recursos novos; quem usava adaptadores arbitrários (GraphQL próprio) não conseguiu migrar ([HN 45389293](https://news.ycombinator.com/item?id=45389293), comentário 45393726).
- **Lacunas enterprise** no lançamento: SCIM ausente impediu adoção ([HN 44030492](https://news.ycombinator.com/item?id=44030492), comentário 44037094); passkey "relegada a plugin" (idem, 44032611). SCIM chegou depois.
- **Issues mais votadas** ([GitHub, ordenado por 👍](https://github.com/better-auth/better-auth/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc)): `useSession()` sem mudança de estado (#1006), `baseURL` dinâmico por requisição para multi-domínio (#4151, aberta), SAML (#96), papéis/permissões dinâmicos (#4557, aberta), multi-tenant de bases de usuários (#1248), audit log (#1184), contas de serviço (#3224, aberta).
- Ritmo de releases e quebra com ORMs (Drizzle v1 #6766, Prisma 8 #9636).
- "Seemed to be run by one guy" (2025) ([HN 48038827](https://news.ycombinator.com/item?id=48038827), comentário 48039514).

## Histórico de segurança
Fonte: OSV `npm:better-auth` (26 advisories, 12/2024–10/2026) e `npm:next-auth`/`@auth/core` (14).

| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVE-2024-56734, CVE-2025-27143, CVE-2025-53535 | 2024–25 | Open redirect (verify-email, callback sem esquema, middleware de origem) | Validação de redirect centralizada, lista exata, testes de parser de URL |
| CVE-2025-71403 | 2025 | Bypass de `trustedOrigins` → ATO | Checagem de origem com parser canônico, nunca prefixo/substring |
| CVE-2025-61928 | 2025 | Criação de API key sem autenticação para qualquer `userId` (crítico; [ZeroPath](https://zeropath.com/blog/breaking-authentication-unauthenticated-api-key-creation-in-better-auth-cve-2025-61928)) | Endpoints de plugin herdam autenticação por padrão; teste "negativo" obrigatório em toda rota |
| CVE-2025-71399 | 2025 | Normalização `//` no roteador burla `disabledPaths` e rate limit | Normalizar caminho antes de qualquer decisão |
| CVE-2026-67337 | 2026 | **Bypass de 2FA** via cache de sessão prematuro | Sessão só nasce depois de todos os fatores; estado "parcial" separado |
| CVE-2026-67335 | 2026 | `state` OAuth não conferido sem PKCE | PKCE + state sempre; nada opcional |
| CVE-2026-45364 | 2026 | Rate limit por IPv6 individual (rotação de prefixo) | Agregar IPv6 por /64 (ou configurável) |
| CVE-2026-45337, CVE-2026-109346 | 2026 | Device flow: aprovar com qualquer sessão; tela não mostra cliente/escopos | Vincular aprovação ao usuário; tela de consentimento informativa |
| CVE-2026-53512, -53517, -53518 | 2026 | Replay de refresh token sem autenticação de cliente; rotação "bifurca" família em concorrência; código de autorização resgatado 2× (find-then-delete) | Operações de uso único atômicas no banco; testes de concorrência |
| CVE-2026-53516, CVE-2026-67327, CVE-2026-53514 | 2026 | ATO por auto-link a e-mail não verificado; pre-account hijacking em magic link/OTP; convite aceito por e-mail não verificado | Confirma a boa prática do brief (5.1): vincular só com e-mail verificado |
| CVE-2026-67336 | 2026 | `alg=none` anunciado e PKCE `plain` aceito por padrão | Padrões seguros por definição (premissa 2) com alerta ao afrouxar |
| CVE-2026-67333 | 2026 | XSS armazenado via `redirect_uri` `javascript:` | Validar esquema de `redirect_uri` no registro do cliente |
| CVE-2026-67334 | 2026 | Sessões sobrevivem à exclusão do usuário (admin, anônimo, SCIM) | Exclusão/desprovisionamento revoga tudo em cascata |
| CVE-2026-110862, -110864 | 2026 | `state` OAuth usado como magic link; OAuth Proxy permite entrar como outro | Separar tipos de token por finalidade (token "tipado") |
| Auth.js CVE-2023-27490, CVE-2026-73419/73420/73421 | 2023–26 | Faltavam checagens de state/nonce/PKCE; cookies não vinculados ao provedor; homóglifo em e-mail; erro de configuração "falha aberto" | Normalizar Unicode antes de validar; falhar fechado em erro de configuração |

Lição geral: a superfície cresce com plugins de protocolo escritos rapidamente; a maioria das falhas é lógica (vinculação, atomicidade, padrões inseguros), não criptográfica.

## Arquitetura e decisões notáveis
- Handler HTTP único montado na aplicação; endpoints por plugin; hooks `before/after`; cliente gerado por inferência de tipos.
- Sessão = cookie opaco + linha em banco; "cookie cache" opcional reduz consultas (e originou o bypass de 2FA).
- Rate limit desligado em desenvolvimento por padrão; armazenamento em memória por padrão (inadequado em serverless, a doc avisa).
- `scrypt` por ser nativo do Node, Argon2id plugável.
- Esquema gerado pela CLI para o ORM do cliente; o esquema é contrato público (mudanças exigem migração do cliente).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy B2B via plugin `organization` (membros, times, papéis, controle de acesso declarado em código, papéis dinâmicos). Bases de usuários isoladas por tenant: não nativo (issue #1248).
- Extensibilidade: plugins de primeira e terceira parte, hooks, `databaseHooks`, campos adicionais no esquema.
- Configuração: **em código** (objeto `betterAuth({...})`), não em runtime/painel — oposto da premissa 3/18 do projeto.
- Bancos: PostgreSQL, MySQL, SQLite/D1, MSSQL, MongoDB (+ Drizzle/Prisma).
- Soberania: total (dados no banco do cliente), exceto se usar o Infrastructure pago.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Só aprovação humana única de agentes/dispositivos (device flow/CIBA no `agent-auth`). Nada em `organization` ou `admin`.
- **Autorização sensível ao estado / transições declarativas**: **não**. Controle de acesso do `organization`/`admin` é RBAC estático (recurso → ações).

## Implicações para o projeto
1. **Adotar** a ideia de "cliente tipado gerado" para nossos SDKs (OpenAPI → SDKs com tipos), que é o principal motivo do elogio de DX.
2. **Adotar** como suíte de regressão as classes de falha acima (auto-link, pre-account hijacking, uso único atômico, IPv6, redirect, `alg=none`, sessão após exclusão) — viram casos no modelo de ameaças e testes ofensivos (4.19).
3. **Evitar** padrões opcionais inseguros (PKCE `plain`, sem PKCE, rate limit off): seguro por padrão + alerta (premissa 4).
4. **Evitar** "sessão completa antes do último fator"; modelar estados de autenticação explicitamente (máquina de estados).
5. **Oportunidade**: público que quer "dados no meu banco" mas teme aquisição/VC — oferecer o mesmo grau de soberania com governança aberta e configuração em runtime.
6. **Oportunidade**: Better Auth deixa detecção de bots, audit log e dashboard no pago; incluí-los no open source é diferencial.
7. **Oportunidade**: modo biblioteca do nosso AuthN (premissa 1) deve ter a mesma ergonomia; Go não tem equivalente popular ([HN 45389293](https://news.ycombinator.com/item?id=45389293): "I really wish there was such an easy off-the shelf auth solution for Go").

## Fontes
- https://github.com/better-auth/better-auth — 2026-11-03
- https://raw.githubusercontent.com/better-auth/better-auth/main/docs/content/docs/ (plugins/sso, 2fa, passkey, oauth-provider, organization, agent-auth, have-i-been-pwned, captcha; concepts/rate-limit, session-management, database; authentication/email-password) — 2026-11-03 — documentação oficial
- https://github.com/better-auth/better-auth/tree/main/docs/content/docs/plugins — 2026-11-03
- https://github.com/better-auth/better-auth/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://api.deps.dev/v3/projects/github.com%2Fbetter-auth%2Fbetter-auth — 2026-11-03
- https://registry.npmjs.org/better-auth e https://api.npmjs.org/downloads/point/last-week/better-auth — 2026-11-03
- https://api.osv.dev (npm:better-auth, npm:next-auth, npm:@auth/core) — 2026-11-03
- https://www.better-auth.com/blog/authjs-joins-better-auth (via busca) — 2026-11-03 — fornecedor
- https://github.com/nextauthjs/next-auth/discussions/13252 — 2026-11-03
- https://vercel.com/blog/vercel-acquires-better-auth (via busca) — 2026-11-03 — comprador
- https://thenewstack.io/vercel-acquires-better-auth/ — 2026-11-03 — imprensa
- https://workos.com/blog/should-i-migrate-from-better-auth-vercel-acquisition — 2026-11-03 — **concorrente**
- https://better-auth.com/docs/infrastructure/introduction e /products (via busca; preços divergentes) — 2026-11-03 — fornecedor
- https://zeropath.com/blog/breaking-authentication-unauthenticated-api-key-creation-in-better-auth-cve-2025-61928 — 2026-11-03 — empresa de segurança (marketing próprio)
- HN: https://news.ycombinator.com/item?id=44030492, https://news.ycombinator.com/item?id=45389293, https://news.ycombinator.com/item?id=48038827, https://news.ycombinator.com/item?id=48819512, https://news.ycombinator.com/item?id=44380185 — 2026-11-03 (API Algolia)
