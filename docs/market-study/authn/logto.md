# Logto

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted (OSS) e SaaS (Logto Cloud); CIAM/B2B SaaS e agentes de IA |
| Site / repositório | https://logto.io · https://github.com/logto-io/logto |
| Licença e modelo de negócio | MPL-2.0 (OSS). Cloud: Free (até 50k MAU, 50k tokens), Pro a partir de US$ 24/mês **cobrado por access token emitido** (US$ 0,08/100 extras) + add-ons de US$ 48 (RBAC, Organizations, MFA, Enterprise SSO cada um, segurança avançada), SAML app US$ 96/app. Self-hosted Pro US$ 199–249/mês (early access) e Enterprise (SLA, LTS). Pago no self-host: colaboradores no console, MFA obrigatório no console, Bring your UI, ocultar marca, IdP-initiated SSO, SAML apps ilimitados (OSS limitado a 3) |
| Stack | TypeScript/Node.js (Node ^26 nos docs atuais), monólito `@logto/core` sobre **node-oidc-provider** (panva, certificado OpenID). PostgreSQL ^14 obrigatório; Redis opcional como cache central. UI em React |
| Maturidade | Repositório criado em jun/2021; Show HN em jul/2022. ~14,8 mil estrelas, ~1,2 mil forks, ~140 issues abertas, push diário. Cloud com SOC 2 (relatório no Pro) |
| Acesso | 2026-11-03 |

> Limitação: GitHub web e parte da API com rate limit; Reddit inacessível. Usei docs.logto.io (sitemap e páginas), logto.io, API do GitHub e HN Algolia.

## Resumo
- Foco em experiência de desenvolvedor: console amigável, ~80 quick-starts, SDKs, sign-in experience pronta e customizável.
- Forte em B2B: Organizations (multi-tenancy com papéis por organização, JIT, convites, MFA por organização) e Enterprise SSO SAML/OIDC.
- Investimento visível em IA: auth para servidores MCP, acesso de agentes de terceiros, Secret Vault de tokens federados.
- OSS cobre a maior parte do produto; monetização por Cloud com add-ons e por licença self-hosted.
- Muitos advisories em 2026, incluindo tomada de conta.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha com política, códigos por e-mail/SMS, links mágicos (convites e login), passkey como primeiro fator, social (Google, Apple, GitHub, Discord etc.), Enterprise SSO SAML e OIDC (SP e IdP-initiated). LDAP/AD: não encontrado nos docs (pedido #1588 fechado). ICP-Brasil: não |
| 2 | MFA e step-up | ◐ | TOTP, WebAuthn, e-mail, SMS, backup codes, trusted devices (2026), step-up/"security verification" para ações sensíveis, MFA obrigatório por organização. MFA adaptativo por risco: não encontrado |
| 3 | Ciclo de vida da conta | ◐ | Cadastro, verificação, reset de senha, coleta de perfil, convites, JIT por organização/domínio, suspensão, migração de usuários (hash legado), Account API/Account Center. SCIM: não encontrado. Verificação de idade: não |
| 4 | Sessões e tokens | ◐ | Gestão de sessões e grants por usuário, back-channel logout, rotação de chaves de assinatura, refresh tokens, opaque vs JWT. DPoP: não encontrado nos docs (o node-oidc-provider suporta, exposição não verificada) |
| 5 | Protocolos | ◐ | OAuth 2.1/OIDC, SAML (apps e SSO), WebAuthn, token exchange. Device flow provável (PR #9832 indexa "user code", 2026), não documentado no sitemap. SCIM, CIBA, FAPI: não encontrados |
| 6 | Máquinas e agentes | ● | M2M (client credentials), personal access tokens, delegação serviço-a-serviço, impersonação por token exchange, auth para MCP e para agentes de IA de terceiros, Secret Vault com tokens federados |
| 7 | Detecção e defesa | ◐ | Captcha (reCAPTCHA Enterprise, Turnstile), identifier lockout, blocklist de e-mail/descartáveis/sub-endereços, rate limit de envio. No Cloud parte é add-on pago. Senhas vazadas e anomalias: não encontrados |
| 8 | Privacidade e conformidade | ◐ | Termos e privacidade na tela de cadastro, exclusão via API, região de dados customizada (Enterprise Cloud), SOC 2 (Cloud), HIPAA/BAA (Enterprise) |
| 9 | Experiência e administração | ● | Console admin, sign-in experience com live preview, CSS customizado, Bring your UI, i18n com CLI de tradução, Account Center pronto, webhooks, "actions" (pós-primeiro fator, pós-login), custom token claims em JS, audit logs |
| 10 | Operação e migração | ◐ | CLI de instalação/upgrade/alteração de banco, rotação de chaves, Docker, migração de usuários. HA com Redis (pedido #1726 fechado). LTS só no plano pago |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Organizations** com template de papéis/permissões por organização, API resources por organização, organization tokens e switcher de organização na UI.
- **Cobrança por token emitido** (não por MAU) no Cloud: modelo de preço incomum e transparente sobre o que conta.
- **Auth para MCP e agentes**: guias para proteger servidores MCP, liberar acesso de agentes de terceiros e conectar agentes a APIs externas com tokens federados guardados no Secret Vault.
- **Experience API**: API de interação para construir UI própria de login/cadastro/MFA (separada da Management API e da Account API).
- OIDC sobre biblioteca certificada (node-oidc-provider) em vez de implementação própria.

## O que a comunidade elogia (com evidência)
- Simplicidade frente ao Keycloak ("Keycloak exige conhecimento profissional… Logto é usável sem conhecer OAuth/OIDC") — HN 32032746 (2022, viés: fundador).
- Show HN com boa recepção (59 e 57 pontos) — HN 32118950, 32031649 (2022).
- Pedido por "rules/actions" estilo Auth0 citado como motivo de não migrar; depois entregue como custom claims e actions — HN 32033658 (2022).

## Do que a comunidade se queixa (com evidência)
- Issues mais reagidas pedem recursos ausentes no início: CAPTCHA (#1751), outros bancos além de Postgres (#1661), LDAP (#1588), SDKs (SvelteKit #3008, Nuxt #1830, Hono #6771 aberto), papéis no JWT (#3411), alta disponibilidade (#1726), deploy em subpath (#1969).
- Paywall: recursos comuns (MFA, RBAC, Organizations, Enterprise SSO) são add-ons pagos no Cloud; no OSS, SAML apps limitados a 3 e "Bring your UI" pago (doc oficial).
- Pedido de página unificada de apps para usuário logado (#5721, aberto desde 2024).

## Histórico de segurança
~15 advisories no GitHub, todos de 2026. Seleção:

| ID | Ano | Classe | Lição |
|---|---|---|---|
| GHSA-p88q-p28c-j9jf (alto) | 2026 | `native_callback` controlado pelo atacante vaza authorization code no fluxo social/SSO → tomada de conta | Todo destino de redirecionamento (inclusive nativo) validado contra registro exato |
| GHSA-crc8-q4hm-4564 (alto) | 2026 | App de terceiro com só `openid` define senha inicial em contas elegíveis | Account API exige escopo e step-up específicos para operações de credencial |
| GHSA-vrv3-3j98-vfpp (alto) | 2026 | Token exchange aceita ID token como subject e não checa o cliente emissor | Token exchange com validação estrita de tipo, audiência e emissor |
| CVE-2026-55377 (alto) | 2026 | Registro WebAuthn contorna step-up do Account Center | Adicionar fator é operação sensível, exige step-up |
| CVE-2026-55370 (médio) | 2026 | Replay de TOTP dentro da janela | Marcar código TOTP usado (anti-replay) |
| CVE-2026-63203 (alto) | 2026 | Account API expõe tokens de provedores sem escopo `identities` | Segredos federados com escopo e auditoria próprios |
| CVE-2026-55789 (alto) | 2026 | Atributos de perfil injetados sem escape em asserções SAML assinadas | Escapar/serializar XML por biblioteca; atributos do usuário são entrada hostil |
| GHSA-5wr4-942j-h43m (alto) | 2026 | JWT Customizer (VM JS) sem timeout efetivo → DoS por tenant | Código de cliente com limites de tempo/memória e isolamento por tenant |
| CVE-2026-56739 (alto) | 2026 | SSRF via webhooks e userinfo de conector OAuth custom | URLs de saída com proteção central contra SSRF |
| CVE-2026-62317, CVE-2026-63188, CVE-2026-54714, CVE-2026-63187 | 2026 | ReDoS, path traversal em `logto-tunnel`, XSS via RelayState, injeção de comando em workflow de CI | Regex com entrada escapada; segurança do CI faz parte do produto (brief 4.18) |

## Arquitetura e decisões notáveis
- Monólito Node com três famílias de API: Management (admin), Experience (UI de login), Account (usuário final), mais OIDC montado em `/oidc`.
- Reuso de biblioteca OIDC certificada em vez de implementação própria.
- Conectores plugáveis (social, e-mail, SMS) como pacotes; CLI para gerenciar conectores e migrações de banco.
- Customização de claims via scripts JS executados em VM no servidor (gerou advisory de DoS).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Dois níveis: **tenant** (pool de usuários isolado; multi-tenant no console só no Cloud) e **organizations** dentro de um tenant (pool único de usuários para apps multi-tenant B2B).
- Extensibilidade: conectores, webhooks, actions, custom claims em JS, Management API.
- Banco: somente PostgreSQL (pedido por outros bancos fechado).
- Soberania: OSS self-hosted "dados nunca saem da sua infraestrutura"; Cloud com região customizada no Enterprise.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** (não encontrado no sitemap de docs).
- **Autorização sensível ao estado / transições**: **não**. Autorização = RBAC global e por organização com permissões em API resources.

## Implicações para o projeto
- **Adotar**: separação de APIs por público (admin, experiência de login, conta do usuário) com escopos distintos.
- **Adotar**: Organizations como modelo de cliente multi-tenant (premissa 10), com template de papéis e MFA por organização.
- **Adotar**: anti-replay de TOTP e step-up obrigatório para cadastrar/remover fatores desde o Marco 1.
- **Evitar**: paywall em segurança básica (MFA, SSO) como fazem os add-ons do Cloud; gera queixa e contraria a premissa 2.
- **Evitar**: scripts de cliente sem sandbox com limites; token exchange sem validação estrita.
- **Oportunidade**: auth para MCP/agentes já é tabela de aposta (Logto, Authgear, Keycloak); entrar com perfil completo (CIMD, DCR, delegação `act`).
- **Oportunidade**: suporte a LDAP/AD e SCIM, ausentes no Logto.

## Fontes
- https://github.com/logto-io/logto — 2026-11-03
- https://api.github.com/repos/logto-io/logto — 2026-11-03
- https://api.github.com/repos/logto-io/logto/security-advisories — 2026-11-03
- https://api.github.com/search/issues?q=repo:logto-io/logto+is:issue&sort=reactions — 2026-11-03
- https://api.github.com/search/issues?q=repo:logto-io/logto+device+flow — 2026-11-03
- https://logto.io/pricing — 2026-11-03 — viés: fornecedor
- https://logto.io/self-hosted-plans — 2026-11-03 — viés: fornecedor
- https://docs.logto.io/sitemap.xml — 2026-11-03
- https://docs.logto.io/logto-oss — 2026-11-03
- https://docs.logto.io/logto-oss/get-started-with-oss — 2026-11-03
- https://docs.logto.io/concepts/core-service — 2026-11-03
- https://docs.logto.io/end-user-flows/mfa — 2026-11-03
- https://news.ycombinator.com/item?id=32031649 , 32118950, 32032746 (viés: fundador), 32033658 — 2026-11-03
