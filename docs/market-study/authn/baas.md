# BaaS: Supabase Auth, Firebase Auth / Identity Platform, Appwrite, PocketBase — comparativo

| Campo | Supabase Auth | Firebase Auth / Identity Platform | Appwrite | PocketBase |
|---|---|---|---|---|
| Categoria | Módulo de auth de BaaS sobre Postgres | Auth gerenciado do Google (Firebase) e sua versão "upgrade" (Identity Platform, GCP) | BaaS completo self-hosted/cloud | BaaS em binário único (Go + SQLite), também usável como framework Go |
| Site / repositório | https://supabase.com/docs/guides/auth · https://github.com/supabase/auth | https://firebase.google.com/docs/auth · https://cloud.google.com/identity-platform | https://appwrite.io · https://github.com/appwrite/appwrite | https://pocketbase.io · repositório **movido** de github.com/pocketbase/pocketbase (arquivado em 26/out/2026) para ganigeorgiev.com/pocketbase/pocketbase |
| Licença e modelo de negócio | MIT (servidor de auth; fork divergente do GoTrue da Netlify). Cloud: Free 50 mil MAU; Pro/Team 100 mil MAU + US$ 0,00325/MAU; SAML SSO 50 MAU incluídos + US$ 0,015/MAU; MFA por telefone US$ 75/mês; proteção de senha vazada, sessão única e timeouts só no Pro+ | Proprietário. Identity Platform: 0–50 mil MAU grátis (tier 1: e-mail, telefone, anônimo, social), depois US$ 0,0055→0,0025/MAU; OIDC/SAML 50 MAU grátis, depois US$ 0,015/MAU; SMS por mensagem (10/dia grátis). Spark sem cartão: 3.000 DAU/dia | BSD-3-Clause; Appwrite Cloud pago por plano | MIT; sem oferta comercial; um mantenedor; financiamento instável (recusou o FLOSS/fund em fev/2026) |
| Stack | Go; **somente PostgreSQL**; integra com RLS via PostgREST | Infra Google; SDKs Web/iOS/Android/Flutter/C++/Unity, Admin SDKs | PHP (Swoole); MariaDB, MongoDB e PostgreSQL (adapters no compose atual), Redis, ClickHouse, Traefik | Go; SQLite embutido; hooks em Go ou JS (VM embutida) |
| Maturidade | ~2,6 mil estrelas no repositório de auth (215 issues, 181 PRs); Supabase entre os BaaS mais adotados | Mais antigo (Firebase Auth anunciado em 2014, [HN 8406887](https://news.ycombinator.com/item?id=8406887)); SLA 99,95% só no Identity Platform | ~57,9 mil estrelas, 288 issues; versão 2.6.x; 1.0 estável em 2022 | ~61,6 mil estrelas; v0.41.1; **pré-1.0** sem garantia de compatibilidade |
| Acesso | 2026-11-03 | 2026-11-03 | 2026-11-03 | 2026-11-03 |

> Limitações: `api.github.com` bloqueado; Firebase é fechado (sem issues públicas). O orçamento de buscas web acabou antes de levantar queixas sobre Firebase (Dynamic Links, fraude de SMS); usei só a documentação oficial para esses pontos.

## Resumo
- Os quatro tratam auth como **parte do backend de dados**: identidade do usuário alimenta regras de acesso a linhas/documentos (RLS no Postgres, Security Rules, permissões de documento, API rules).
- **Supabase** é o mais rico em padrões: OAuth 2.1/OIDC server, SAML SP, passkeys (experimental), hooks, rotação de refresh com detecção de reuso.
- **Firebase** é o mais maduro e escalável, mas MFA, SAML/OIDC, multi-tenancy e blocking functions só no Identity Platform; sem passkeys nativas.
- **Appwrite** tem o pacote de ciclo de vida mais completo (equipes, convites, impersonação auditável, políticas de senha/e-mail, passkeys), mas sem SAML.
- **PocketBase** é minimalista: tokens sem estado, sem logout no servidor, MFA "quaisquer 2 métodos"; risco de projeto de um mantenedor só.

## Funcionalidades por domínio (5.1)

| # | Domínio | Supabase | Firebase / IdP | Appwrite | PocketBase |
|---|---|---|---|---|---|
| 1 | Autenticação primária | ◐ Senha, link mágico, OTP por e-mail, telefone (Twilio, MessageBird, Vonage), 19 sociais + OAuth/OIDC customizado, SAML 2.0 SSO, anônimo, Web3 (Sign in with Ethereum), passkeys **experimentais**. Sem LDAP/ICP | ◐ Senha, link por e-mail, telefone, anônimo, sociais (Google, Apple, Facebook, GitHub, Microsoft, X, Yahoo...), custom tokens; SAML/OIDC só no IdP. Sem passkeys nativas (sem página na doc; extensão comunitária) | ◐ Senha (Argon2), Magic URL, OTP por e-mail, SMS, 30+ OAuth2, anônimo, **passkeys** (2.5+), JWT, custom token. Sem SAML/LDAP | ◐ Senha (identidade configurável), OTP por e-mail, OAuth2 (Google, GitHub, Microsoft...). Sem passkeys, SAML, telefone |
| 2 | MFA e step-up | ◐ TOTP; telefone (SMS/WhatsApp, pago); WebAuthn como 2º fator experimental; níveis AAL1/AAL2 no JWT. Adaptativo: não | ◐ SMS e TOTP (IdP). Adaptativo: não | ◐ TOTP, e-mail, telefone, códigos de recuperação; exigir MFA em ações sensíveis | ◐ MFA = autenticar com 2 métodos diferentes quaisquer (ex.: senha + OTP); sem TOTP |
| 3 | Ciclo de vida da conta | ◐ Cadastro, verificação, recuperação, vinculação manual/automática, convite por admin, ban, exclusão; hook "before user created". Sem SCIM | ◐ Cadastro, verificação, linking, enumeração protegida por padrão (projetos desde 15/set/2023), blocking functions (IdP), importação com vários hashes. Sem SCIM | ● Contas, verificação, recuperação, **equipes com convites e papéis**, labels, preferências, políticas de e-mail (bloquear descartáveis/aliases/gratuitos). Sem SCIM | ◐ Verificação, reset, troca de e-mail; múltiplas coleções de auth (users, staff...). Sem convites/SCIM |
| 4 | Sessões e tokens | ◐ JWT curto + refresh com rotação e **detecção de reuso** (janela 10 s; reuso fora dela revoga a sessão), logout global/local/outros, sessão única por usuário e timeouts (Pro+), chaves de assinatura assimétricas. Back-channel/DPoP: não | ◐ ID token + refresh; revogação de refresh tokens via Admin SDK; session cookies para servidor | ◐ Duração, limite de sessões (mais antiga é removida), alertas de sessão, invalidação ao trocar senha | ○ Tokens **sem estado e sem logout**; revogação só trocando a senha ou o segredo `tokenKey` da coleção |
| 5 | Protocolos | ◐ **OAuth 2.1 + OIDC server** (sem custo extra), SAML SP. Sem SCIM, CIBA, PAR, FAPI | ◐ RP OIDC/SAML (IdP); não é provedor OIDC | ◐ "OAuth2 server / Sign in with Appwrite" (nav da doc). Sem SAML | ○ Só cliente OAuth2 |
| 6 | Máquinas e agentes | ◐ Chaves de projeto (publishable/secret); autenticação de agentes MCP via OAuth 2.1 server; third-party auth (Clerk, Auth0, Firebase, Cognito, WorkOS) | ◐ Service accounts do GCP, custom tokens | ◐ Chaves de API de projeto, JWT | ◐ Não tem chaves de API: recomenda token de impersonação de superusuário (não renovável) |
| 7 | Detecção e defesa | ◐ Rate limits, CAPTCHA, **senhas vazadas (HIBP)** no Pro+ | ◐ reCAPTCHA Enterprise com avaliação de **fraude de SMS (toll fraud)** em modo audit/enforce, regiões de SMS, políticas de senha, cotas (100 contas/h por IP) | ◐ Histórico e dicionário de senhas, bloqueio de dados pessoais na senha, rate limits | ◐ Rate limiter embutido (v0.23+), allowlist de IPs para superusuários (v0.38+) |
| 8 | Privacidade e conformidade | ◐ Logs de auditoria (1 h a 28 dias por plano), "user data ownership", self-host | ◐ ISO 27001, SOC 1/2/3; BAA, TISAX e PCI-DSS só no IdP | ◐ Self-host; impersonação marcada (`impersonatorUserId`) | ○ Nada específico |
| 9 | Experiência e administração | ◐ Dashboard; sem tela de login hospedada; templates de e-mail sem i18n (#80, aberta desde 2021) | ◐ FirebaseUI drop-in; console | ● Console, SSR (Next/Nuxt/SvelteKit), **impersonação de usuário** controlada, presença, MCP/CLI/Terraform | ◐ Dashboard admin, templates de e-mail; impersonação por superusuário |
| 10 | Operação e migração | ◐ Self-host (Docker); Postgres | ◐ SaaS; import/export de usuários com hashes (bcrypt, scrypt, Argon2, PBKDF2, HMAC) | ◐ Docker/Helm, Coolify; migrações | ◐ Binário único, systemd; backup do SQLite |
| 11 | Identidade descentralizada | ○ (Web3 é login por carteira, não VC) | ○ | ○ | ○ |

## Diferenciais / exclusivos
- **Supabase**: hooks em funções Postgres (custom access token, MFA verification, password verification, send email/SMS, before user created); OAuth 2.1 server para "Sign in with [Seu App]" e MCP; detecção de reuso de refresh token documentada; aceitar JWT de IdPs terceiros nas mesmas regras RLS.
- **Firebase/IdP**: **multi-tenancy** nativa (tenants com IdPs próprios num projeto, voltado a B2B — "the best I've ever seen", [HN 41926940](https://news.ycombinator.com/item?id=41926940)); defesa contra fraude de SMS integrada; enumeração de e-mail protegida por padrão; SLA 99,95%.
- **Appwrite**: impersonação com marca no modelo do usuário; políticas de e-mail; equipes como unidade de multi-tenancy; presença em tempo real.
- **PocketBase**: MFA genérico "dois métodos quaisquer"; `@request.context` nas regras (distingue `oauth2`, `otp`, `password`...); tudo num binário.

## O que a comunidade elogia (com evidência)
- **PocketBase**: simplicidade "one file", SSE, fácil para iniciantes; repetidamente no topo do HN (563, 630 e 671 pontos em 2022, 2024 e 2025: [HN 32013330](https://news.ycombinator.com/item?id=32013330), [HN 38898934](https://news.ycombinator.com/item?id=38898934), [HN 46075320](https://news.ycombinator.com/item?id=46075320)); "made my life so much easier" ([HN 41926038](https://news.ycombinator.com/item?id=41926038)).
- **Supabase**: lançamento de SSO/mobile/server-side bem recebido ([HN 35555263](https://news.ycombinator.com/item?id=35555263), 183 pts); anônimo era o pedido nº 1 (#68, 142 reações) e foi entregue em 2024.
- **Appwrite**: "open source Firebase" self-hosted ([HN 30769044](https://news.ycombinator.com/item?id=30769044), 326 pts).
- **Firebase**: tenants e escala; experiência inicial rápida.

## Do que a comunidade se queixa (com evidência)
- **Firebase**: "consistently lead to misery", "No one should use Firebase today", "Worst auth experience: Firebase on NextJS" ([HN 41925617](https://news.ycombinator.com/item?id=41925617), [41925864](https://news.ycombinator.com/item?id=41925864), [41926038](https://news.ycombinator.com/item?id=41926038)); recursos básicos (MFA, SAML) só no upgrade pago; o fim do Firebase Dynamic Links (25/ago/2025) quebrou link por e-mail em apps móveis e OAuth em SDKs antigos, exigindo migração (doc oficial).
- **Supabase** ([issues por reações](https://github.com/supabase/auth/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc)): `signUp()` não falha para conta existente (#1517, 98 — efeito da proteção contra enumeração, confunde devs); i18n de templates de e-mail (#80, 62, desde 2021); vincular vários provedores (#313, 42); códigos de erro (#915); verificação de e-mail sem bloquear login (#2513); passkeys pedidas de 2021 até 2026 (#92). Recursos de segurança (senha vazada, sessão única) atrás do plano Pro.
- **PocketBase**: só SQLite ("if only it supported Postgres", [HN 47063903](https://news.ycombinator.com/item?id=47063903)); risco de "single point of failure" de um mantenedor ([HN 47066792](https://news.ycombinator.com/item?id=47066792); "Will PocketBase Survive?", [HN 42496775](https://news.ycombinator.com/item?id=42496775)); pré-1.0 com quebras; repositório arquivado e movido para fora do GitHub (out/2026).
- **Appwrite**: complexidade operacional (muitos containers: MariaDB/Mongo/Postgres, Redis, ClickHouse, executores; compose atual); queixas públicas específicas de auth: não encontradas.

## Histórico de segurança
| Produto | ID | Ano | Classe |
|---|---|---|---|
| Supabase | GHSA-v36f-qvww-8w8m / CVE-2026-31813 | 2026 | Login por ID token Apple/Azure aceitava **issuer controlado pelo atacante** e vinculava a identidade da vítima → sessão AAL1 |
| Supabase | GHSA-3529-5m8x-rpv3 | 2024 | **Envenenamento de link por e-mail**: links montados a partir de `X-Forwarded-Host/Proto` sem validação → takeover (CVSS 7.3); corrigido com allowlist (vazia por padrão) |
| Supabase | GHSA-9wqv-qhrh-87f4 | 2023 | Fluxo OIDC inseguro |
| Supabase | GHSA-5hvv-9cqv-894r | 2022 | Falha de permissionamento de usuário |
| Supabase (ecossistema) | CVE-2025-48757 (Lovable) | 2025 | RLS ausente/fraca em apps gerados: ~170 apps com tabelas legíveis/graváveis só com a anon key (pesquisa de Matt Palmer) |
| PocketBase | GHSA-pq7p-mc74-g65w | 2026 | **Pre-hijacking** de conta via auto-vinculação OAuth2 de e-mail não verificado → verificado |
| PocketBase | GHSA-m93w-4fxv-r35v | 2024 | Vinculação de senha e OAuth2 com e-mail não verificado |
| PocketBase | GHSA-84vh-m24q-wjjx | 2026 | Panic não tratado em goroutines (DoS) |
| Appwrite | CVE-2023-50974 | 2023 | CLI gravava credenciais com permissões amplas; demais CVEs (SSRF, traversal, XSS) fora de auth; nenhum advisory publicado no GitHub |
| Firebase | (sem CVE de produto) | 2024–2025 | Configurações erradas de Security Rules expuseram ~19,8 mi de credenciais em texto puro (916 sites) e endpoints de apps móveis sem auth (pesquisas e imprensa) |

Lições para nós:
- **Nunca derivar URLs de e-mail de cabeçalhos da requisição**: URL base vem da configuração do tenant/aplicação.
- **Issuer de ID token é configuração, não entrada**: lista fixa de emissores por provedor; testes negativos com emissores falsos.
- **Vinculação automática só com e-mail verificado pelos dois lados** (confirma a [BOA PRÁTICA] do brief); pre-hijacking aparece em dois produtos.
- **Padrões seguros nas regras de dados**: o maior dano real no BaaS vem de regras ausentes (RLS, Security Rules), não de bugs do produto. Negar por padrão e alertar quando uma tabela/coleção fica aberta (premissa 4).
- **Logout e revogação no servidor são obrigatórios**: tokens sem estado (PocketBase) não permitem resposta a incidente.

## Arquitetura e decisões notáveis
- **Supabase**: JWT emitido pelo Auth é consumido pelo Postgres (claims em `auth.jwt()` nas políticas RLS); hooks executam funções SQL/HTTP em pontos do ciclo; AAL no token permite políticas que exigem MFA.
- **Firebase**: ID token curto + refresh; regras (Firestore/Storage/RTDB) recebem `request.auth`; Identity Platform é o backend comum dos dois produtos.
- **Appwrite**: permissões por documento/arquivo com papéis (`user:`, `team:`, `role:`); sessões por cookie com SSR.
- **PocketBase**: regras por ação (`listRule`, `viewRule`, `createRule`, `updateRule`, `deleteRule`, `manageRule`) em linguagem de filtro sobre campos do registro, `@request.auth`, `@request.body`, `@request.context`; superusuários ignoram regras.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: Supabase — um projeto = um pool de usuários (SSO por domínio); Firebase — **tenants** no Identity Platform; Appwrite — equipes; PocketBase — várias coleções de auth, sem tenants.
- **Extensibilidade**: hooks Postgres (Supabase), blocking functions (IdP), Functions (Appwrite), hooks Go/JS (PocketBase).
- **Configuração**: dashboard + CLI/config em todos; Appwrite e Supabase com Terraform/CLI.
- **Bancos**: Postgres (Supabase); proprietário (Firebase); MariaDB/MongoDB/Postgres (Appwrite); SQLite (PocketBase).
- **Soberania**: Supabase, Appwrite e PocketBase self-hosted; Firebase só na nuvem Google.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não** nos quatro.
- Autorização sensível ao estado / transições declarativas: **parcial** (no motor de dados, não no auth):
  - PocketBase: regras veem campos atuais do registro e `@request.body.*`, o que permite expressar transições (ex.: atualizar só se o estado atual for X e o novo for Y) — declarativo, mas sem modelo de estados nem definição compartilhada.
  - Firebase Security Rules (Firestore): `resource.data` (antes) e `request.resource.data` (depois) permitem regras de transição; fora do produto de auth.
  - Supabase: RLS com `USING`/`WITH CHECK` sobre a linha; comparar antes/depois exige trigger. Appwrite: não.
  - Nenhum oferece máquina de estados declarativa compartilhada com a aplicação — o diferencial do projeto continua válido.

## Implicações para o projeto
1. **Adotar** como mínimo esperado: senha, link mágico/OTP, social, anônimo convertível, TOTP, rotação de refresh com detecção de reuso, proteção contra enumeração, hooks em pontos do ciclo de vida.
2. **Adotar** defesa contra SMS pumping nativa (orçamento, regiões permitidas, avaliação de risco) antes de habilitar SMS/WhatsApp (5.1, restrições práticas).
3. **Adotar** multi-tenancy no estilo tenants do Identity Platform (IdPs e políticas por tenant) — elogiado e raro.
4. **Adotar** impersonação auditável com marca no token/modelo (Appwrite), sempre registrada (boa prática do brief).
5. **Evitar** recursos de segurança só em plano pago (senha vazada, sessão única, MFA): queixa recorrente; manter no núcleo aberto.
6. **Oportunidade**: transições de estado declarativas no AuthZ integradas ao AuthN — os BaaS mostram demanda (regras antes/depois), mas sem modelo de estados.
7. **Oportunidade**: alertas de "dados abertos" (equivalente a RLS ausente) quando o AuthZ for consumido por apps geradas por IA.

## Fontes
- https://github.com/supabase/auth — 2026-11-03
- https://github.com/supabase/auth/security/advisories (+ GHSA-3529-5m8x-rpv3, GHSA-v36f-qvww-8w8m) — 2026-11-03
- https://github.com/supabase/auth/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://supabase.com/docs/guides/auth.md e https://supabase.com/llms-full.txt (preços, OAuth 2.1 server, passkeys, MFA, hooks, sessões) — 2026-11-03 — fornecedor
- https://supabase.com/docs/guides/auth/sessions.md — 2026-11-03
- https://www.superblocks.com/blog/lovable-rls-breach , https://cvemon.intruder.io/cves/CVE-2025-48757 (via busca) — 2026-11-03 — fornecedor de ferramenta concorrente / agregador
- https://firebase.google.com/docs/auth — 2026-11-03
- https://firebase.google.com/docs/auth/web/start (lista de métodos; ausência de passkeys) — 2026-11-03
- https://firebase.google.com/docs/auth/web/email-link-auth (aviso de Dynamic Links) — 2026-11-03
- https://firebase.google.com/docs/auth/limits — 2026-11-03
- https://cloud.google.com/identity-platform/docs/product-comparison — 2026-11-03
- https://cloud.google.com/identity-platform/pricing — 2026-11-03
- https://cloud.google.com/identity-platform/docs/admin/email-enumeration-protection — 2026-11-03
- https://cloud.google.com/identity-platform/docs/recaptcha-enterprise — 2026-11-03
- https://cloud.google.com/identity-platform/docs/multi-tenancy — 2026-11-03
- https://blog.gitguardian.com/misconfigurations-in-google-firebase-lead-to-over-19-8-million-leaked-secrets , https://cybersecuritynews.com/numerous-applications-using-googles-firebase-platform/ (via busca) — 2026-11-03 — fornecedor de segurança / imprensa
- https://github.com/gavinsawyer/firebase-web-authn (via busca) — 2026-11-03 — comunitário
- https://github.com/appwrite/appwrite e /security/advisories — 2026-11-03
- https://raw.githubusercontent.com/appwrite/appwrite/main/docker-compose.yml e .env — 2026-11-03
- https://appwrite.io/products/auth — 2026-11-03 — fornecedor
- https://appwrite.io/docs/products/auth (passkeys, impersonation, mfa, security) — 2026-11-03
- https://www.wiz.io/vulnerability-database/cve/cve-2023-50974 (via busca) — 2026-11-03 — agregador
- https://raw.githubusercontent.com/pocketbase/pocketbase/master/README.md — 2026-11-03
- https://github.com/pocketbase/pocketbase e /security/advisories — 2026-11-03
- https://pocketbase.io/docs/authentication/ , /docs/collections/ , /docs/api-rules-and-filters/ , /docs/going-to-production/ — 2026-11-03
- HN via API Algolia: 41923641 (e comentários), 47062561, 46075320, 38898934, 32013330, 42496775, 35555263, 30769044, 8406887 — 2026-11-03
